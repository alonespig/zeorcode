package service

import (
	"testing"

	"zoj/internal/model"
)

func TestBuildProblemSetRankCellsKeepsProblemOrderAndStatus(t *testing.T) {
	statuses := map[problemSetRankStatusKey]int{
		{UserID: 9, ProblemID: 101}: model.UserProblemSolved,
		{UserID: 9, ProblemID: 103}: 2,
	}
	cells := buildProblemSetRankCells(
		9,
		[]int64{103, 101, 102},
		map[int64]string{101: "L101", 102: "L102", 103: "L103"},
		statuses,
	)

	if len(cells) != 3 {
		t.Fatalf("len(cells) = %d, want 3", len(cells))
	}
	if cells[0].ProblemID != "L103" || cells[0].Status == nil || *cells[0].Status != 2 {
		t.Fatalf("cells[0] = %+v, want attempted L103", cells[0])
	}
	if cells[1].ProblemID != "L101" || cells[1].Status == nil || *cells[1].Status != model.UserProblemSolved {
		t.Fatalf("cells[1] = %+v, want solved L101", cells[1])
	}
	if cells[2].ProblemID != "L102" || cells[2].Status != nil {
		t.Fatalf("cells[2] = %+v, want untried L102", cells[2])
	}
}
