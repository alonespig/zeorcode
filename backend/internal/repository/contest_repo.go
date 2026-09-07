package repository

import (
	"context"
	"time"

	"zoj/internal/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type ContestRepo struct {
	db *gorm.DB
}

const contestEndTimeExpr = "COALESCE(end_time, DATE_ADD(start_time, INTERVAL duration MINUTE))"

func NewContestRepo(db *gorm.DB) *ContestRepo {
	return &ContestRepo{db: db}
}

// ReplaceContestUserProblems 事务内「先删后插」原子重建某比赛的全部 UserContestProblem。
// 重算时用：整场从 submissions 明细重新生成聚合，保证没有旧残留行。
func (r *ContestRepo) ReplaceContestUserProblems(ctx context.Context, contestID int64, ucps []model.UserContestProblem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("contest_id = ?", contestID).Delete(&model.UserContestProblem{}).Error; err != nil {
			return err
		}
		if len(ucps) == 0 {
			return nil
		}
		return tx.CreateInBatches(ucps, 200).Error
	})
}

// CreateContestWithProblems 原子创建比赛及其题目关联。
func (r *ContestRepo) CreateContestWithProblems(ctx context.Context, contest *model.Contest, problems []model.ContestProblem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(contest).Error; err != nil {
			return err
		}
		if len(problems) == 0 {
			return nil
		}
		for i := range problems {
			problems[i].ContestID = contest.ID
		}
		return tx.Create(&problems).Error
	})
}

// UpdateContestWithProblems 原子更新比赛，并用 problems 完整替换原题目关联。
func (r *ContestRepo) UpdateContestWithProblems(ctx context.Context, contest *model.Contest, problems []model.ContestProblem) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Updates(contest).Error; err != nil {
			return err
		}
		if err := tx.Where("contest_id = ?", contest.ID).Delete(&model.ContestProblem{}).Error; err != nil {
			return err
		}
		if len(problems) == 0 {
			return nil
		}
		for i := range problems {
			problems[i].ContestID = contest.ID
		}
		return tx.Create(&problems).Error
	})
}

func (r *ContestRepo) GetContestByID(ctx context.Context, id int64) (*model.Contest, error) {
	var contest model.Contest
	if err := r.db.WithContext(ctx).First(&contest, id).Error; err != nil {
		return nil, err
	}
	return &contest, nil
}

func (r *ContestRepo) PublicIDExists(ctx context.Context, publicID int64) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&model.Contest{}).
		Where("public_id = ?", publicID).Count(&count).Error
	return count > 0, err
}

func (r *ContestRepo) ResolveIDByPublicID(ctx context.Context, publicID int64) (int64, error) {
	var contest model.Contest
	err := r.db.WithContext(ctx).Select("id").
		Where("public_id = ?", publicID).First(&contest).Error
	return contest.ID, err
}

// GetRatingChangesByContest 取某场比赛所有人的 rating 变化（结算后榜单展示 ±delta 用）。
func (r *ContestRepo) GetRatingChangesByContest(ctx context.Context, contestID int64) ([]model.RatingChange, error) {
	var changes []model.RatingChange
	err := r.db.WithContext(ctx).Where("contest_id = ?", contestID).Find(&changes).Error
	return changes, err
}

// SettleTx 事务内锁定比赛行（FOR UPDATE）后执行 fn，用于 rating 惰性结算的幂等落库。
func (r *ContestRepo) SettleTx(ctx context.Context, contestID int64,
	fn func(tx *gorm.DB, contest *model.Contest) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var contest model.Contest
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&contest, contestID).Error; err != nil {
			return err
		}
		return fn(tx, &contest)
	})
}

func (r *ContestRepo) ListContests(ctx context.Context, page, pageSize int, keyword string, ctype int, status *int) ([]model.Contest, int64, error) {
	var contests []model.Contest
	var total int64

	q := r.db.WithContext(ctx).Model(&model.Contest{})
	if keyword != "" {
		q = q.Where("name LIKE ?", "%"+keyword+"%")
	}
	if ctype > 0 {
		q = q.Where("type = ?", ctype)
	}
	if status != nil {
		q = applyContestStatusFilter(q, *status, time.Now())
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("start_time DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&contests).Error; err != nil {
		return nil, 0, err
	}
	return contests, total, nil
}

func applyContestStatusFilter(q *gorm.DB, status int, now time.Time) *gorm.DB {
	// end_time 是当前比赛结束时间的准确信息；旧数据若未写 end_time，
	// 再按 start_time + duration（分钟）计算，避免状态筛选漏掉历史记录。
	switch status {
	case 0: // 未开始
		return q.Where("start_time > ?", now)
	case 1: // 进行中
		return q.Where("start_time <= ? AND "+contestEndTimeExpr+" > ?", now, now)
	case 2: // 已结束
		return q.Where(contestEndTimeExpr+" <= ?", now)
	default:
		return q
	}
}

func (r *ContestRepo) CreateContestUser(ctx context.Context, userID, contestID int64) error {
	contestUser := &model.ContestUser{
		UserID:    userID,
		ContestID: contestID,
	}
	return r.db.WithContext(ctx).Create(contestUser).Error
}

func (r *ContestRepo) GetContestUsersCount(ctx context.Context, contestID int64) (int64, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ContestUser{}).Where("contest_id = ?", contestID).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}

type ContestUserCount struct {
	ContestID int64 `gorm:"column:contest_id"`
	Count     int64 `gorm:"column:count"`
}

func (r *ContestRepo) GetContestUserCounts(ctx context.Context, contestIDs []int64) (map[int64]int64, error) {
	result := make(map[int64]int64, len(contestIDs))
	if len(contestIDs) == 0 {
		return result, nil
	}
	var rows []ContestUserCount
	err := r.db.WithContext(ctx).Model(&model.ContestUser{}).
		Select("contest_id, COUNT(*) AS count").
		Where("contest_id IN ?", contestIDs).
		Group("contest_id").
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ContestID] = row.Count
	}
	return result, nil
}

func (r *ContestRepo) ExistByID(ctx context.Context, contestID, userID int64) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&model.ContestUser{}).Where("contest_id = ? AND user_id = ?", contestID, userID).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *ContestRepo) GetRegisteredContestIDs(ctx context.Context, userID int64, contestIDs []int64) (map[int64]bool, error) {
	result := make(map[int64]bool, len(contestIDs))
	if userID <= 0 || len(contestIDs) == 0 {
		return result, nil
	}
	var rows []model.ContestUser
	err := r.db.WithContext(ctx).Model(&model.ContestUser{}).
		Select("contest_id").
		Where("user_id = ? AND contest_id IN ?", userID, contestIDs).
		Find(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		result[row.ContestID] = true
	}
	return result, nil
}

func (r *ContestRepo) GetContestProblems(ctx context.Context, contestID int64) ([]model.ContestProblem, error) {
	var problems []model.ContestProblem
	if err := r.db.WithContext(ctx).Where("contest_id = ?", contestID).Find(&problems).Error; err != nil {
		return nil, err
	}
	return problems, nil
}

func (r *ContestRepo) GetContestProblem(ctx context.Context, p model.ContestProblem) (*model.ContestProblem, error) {
	var problem model.ContestProblem
	if err := r.db.WithContext(ctx).Where(&p).First(&problem).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *ContestRepo) GetContestProblemsByContestID(ctx context.Context, contestID int64) ([]model.ContestProblem, error) {
	var problems []model.ContestProblem
	if err := r.db.WithContext(ctx).Where("contest_id = ?", contestID).Find(&problems).Error; err != nil {
		return nil, err
	}
	return problems, nil
}

func (r *ContestRepo) GetContestProblemByID(ctx context.Context, id int64) (*model.ContestProblem, error) {
	var problem model.ContestProblem
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&problem).Error; err != nil {
		return nil, err
	}
	return &problem, nil
}

func (r *ContestRepo) GetContestUserByContestID(ctx context.Context, contestID int64) ([]model.ContestUser, error) {
	var users []model.ContestUser
	if err := r.db.WithContext(ctx).Where("contest_id = ?", contestID).Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}

// GetUserContestProblemsByContestID loads all user/problem result records once.
// The rank service builds userID -> problemID maps from this instead of querying
// inside the nested user/problem loop.
func (r *ContestRepo) GetUserContestProblemsByContestID(ctx context.Context, contestID int64) ([]model.UserContestProblem, error) {
	var records []model.UserContestProblem
	err := r.db.WithContext(ctx).Where("contest_id = ?", contestID).Find(&records).Error
	return records, err
}

// GetUserContestProblem 查询用户在某场比赛某道题的做题记录
func (r *ContestRepo) GetUserContestProblem(ctx context.Context, contestID, userID, problemID int64) (*model.UserContestProblem, error) {
	var record model.UserContestProblem
	err := r.db.WithContext(ctx).Model(&model.UserContestProblem{}).
		Where("contest_id = ? AND user_id = ? AND problem_id = ?",
			contestID, userID, problemID).
		First(&record).Error
	if err != nil {
		return nil, err
	}
	return &record, nil
}
