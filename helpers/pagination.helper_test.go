package sdk_helper

import (
	"math"
	"testing"
)

func TestPagination(t *testing.T) {
	p, err := Pagination(10, 0, 95)
	if err != nil {
		t.Fatal(err)
	}

	if p.Limit != 10 || p.Page != 0 || p.TotalData != 95 {
		t.Fatalf("unexpected %+v", p)
	}

	if p.TotalPage != math.Ceil(9.5) {
		t.Fatalf("TotalPage got %v", p.TotalPage)
	}

	p, _ = Pagination(10, 0, 0)
	if p.TotalPage != 0 {
		t.Fatalf("empty dataset should have 0 pages, got %v", p.TotalPage)
	}

	if _, err := Pagination(0, 0, 10); err == nil {
		t.Fatal("expected error for zero limit")
	}

	if _, err := Pagination(10, -1, 10); err == nil {
		t.Fatal("expected error for negative offset")
	}

	if _, err := Pagination(10, 0, -1); err == nil {
		t.Fatal("expected error for negative total")
	}
}

func TestCalculateOffsetAndPage(t *testing.T) {
	off, err := CalculateOffset(3, 10)
	if err != nil || off != 20 {
		t.Fatalf("offset %d %v", off, err)
	}

	page, err := CalculatePage(20, 10)
	if err != nil || page != 3 {
		t.Fatalf("page %d %v", page, err)
	}

	if _, err := CalculateOffset(0, 10); err == nil {
		t.Fatal("expected error for page 0")
	}

	if _, err := CalculateOffset(1, 0); err == nil {
		t.Fatal("expected error for pageSize 0")
	}

	if _, err := CalculatePage(-1, 10); err == nil {
		t.Fatal("expected error for negative offset")
	}

	if _, err := CalculatePage(10, 0); err == nil {
		t.Fatal("expected error for pageSize 0")
	}

	// Round trip consistency.
	for _, page := range []int{1, 2, 17} {
		o, _ := CalculateOffset(page, 25)
		back, _ := CalculatePage(o, 25)
		if back != page {
			t.Fatalf("round trip failed for page %d", page)
		}
	}
}
