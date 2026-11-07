package cmd

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/alonespig/zeorcode/backend/internal/infra/database"
	"github.com/alonespig/zeorcode/backend/internal/model"
	"github.com/alonespig/zeorcode/backend/internal/repository"
	"github.com/alonespig/zeorcode/backend/pkg/publicid"

	"github.com/spf13/cobra"
	"golang.org/x/crypto/bcrypt"
)

var (
	adminInitUsername string
	adminInitRealName string
	adminInitEmail    string
)

var adminCmd = &cobra.Command{
	Use:   "admin",
	Short: "管理命令",
}

var adminInitCmd = &cobra.Command{
	Use:          "init",
	Short:        "创建初始超级管理员（数据库中没有超管时可用）",
	SilenceUsage: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		db, closeDB := database.NewMysqlClient()
		defer closeDB()
		repo := repository.NewUserRepo(db)

		count, err := repo.CountByRole(ctx, model.RoleSuperAdmin)
		if err != nil {
			return fmt.Errorf("检查超级管理员失败: %w", err)
		}
		if count > 0 {
			return errors.New("系统中已存在超级管理员，拒绝重复初始化")
		}

		username := strings.TrimSpace(adminInitUsername)
		if username == "" {
			return errors.New("用户名不能为空")
		}
		realName := strings.TrimSpace(adminInitRealName)
		if realName == "" {
			realName = "系统管理员"
		}

		password := os.Getenv("ZEORCODE_INITIAL_ADMIN_PASSWORD")
		if password == "" {
			fmt.Print("请输入超级管理员密码（交互输入时内容会显示，至少 8 位）：")
			line, readErr := bufio.NewReader(os.Stdin).ReadString('\n')
			if readErr != nil {
				return fmt.Errorf("读取密码失败: %w", readErr)
			}
			password = strings.TrimSpace(line)
		}
		if len(password) < 8 {
			return errors.New("密码至少 8 位")
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return fmt.Errorf("生成密码哈希失败: %w", err)
		}
		uid, err := allocateUID(ctx, repo)
		if err != nil {
			return err
		}

		user := &model.User{
			UID:      uid,
			Role:     model.RoleSuperAdmin,
			Username: username,
			RealName: realName,
			Password: string(hash),
			Status:   model.UserStatusNormal,
		}
		if email := strings.TrimSpace(adminInitEmail); email != "" {
			user.Email = &email
		}
		if err := repo.CreateUser(ctx, user); err != nil {
			return fmt.Errorf("创建超级管理员失败: %w", err)
		}
		fmt.Printf("已创建超级管理员：%s（用户号 %d）\n", username, uid)
		return nil
	},
}

func allocateUID(ctx context.Context, repo *repository.UserRepo) (int64, error) {
	for i := 0; i < 20; i++ {
		uid, err := publicid.New()
		if err != nil {
			return 0, err
		}
		exists, err := repo.ExistsByUID(ctx, uid)
		if err != nil {
			return 0, fmt.Errorf("检查用户号失败: %w", err)
		}
		if !exists {
			return uid, nil
		}
	}
	return 0, errors.New("生成用户号失败，请重试")
}

func init() {
	adminInitCmd.Flags().StringVar(&adminInitUsername, "username", "admin", "登录用户名")
	adminInitCmd.Flags().StringVar(&adminInitRealName, "real-name", "系统管理员", "真实姓名")
	adminInitCmd.Flags().StringVar(&adminInitEmail, "email", "", "邮箱（可选）")
	adminCmd.AddCommand(adminInitCmd)
	RootCmd.AddCommand(adminCmd)
}
