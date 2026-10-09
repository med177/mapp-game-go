package ui

import "testing"

func TestListViewContentInsetsShareDrawAndHitGeometry(t *testing.T) {
	list := NewListView(10, 20, 200, 120, 24, 4, []string{"one", "two", "three"})
	list.ContentInsets = ListViewInsets{Top: 8, Right: 10, Bottom: 12, Left: 6}

	wantContent := Rect{X: 16, Y: 28, W: 184, H: 100}
	if got := list.ContentRect(); got != wantContent {
		t.Fatalf("ContentRect() = %+v, want %+v", got, wantContent)
	}
	if got := list.ItemIndexAt(20, 20); got != -1 {
		t.Fatalf("ItemIndexAt() in top inset = %d, want -1", got)
	}
	if got := list.ItemIndexAt(20, 29); got != 0 {
		t.Fatalf("ItemIndexAt() in first row = %d, want 0", got)
	}
	if got := list.ItemIndexAt(20, 53); got != 1 {
		t.Fatalf("ItemIndexAt() in second row = %d, want 1", got)
	}
	if got := list.ItemIndexAt(20, 128); got != -1 {
		t.Fatalf("ItemIndexAt() outside outer bounds = %d, want -1", got)
	}
}

func TestListViewItemIndexAtCountDoesNotRequireLabels(t *testing.T) {
	list := ListView{
		Rect:          Rect{X: 0, Y: 0, W: 100, H: 100},
		ContentInsets: ListViewInsets{Top: 10},
		RowHeight:     20,
		VisibleRows:   3,
		Scroll:        1,
	}
	if got := list.ItemIndexAtCount(20, 5, 4); got != -1 {
		t.Fatalf("ItemIndexAtCount() in top inset = %d, want -1", got)
	}
	if got := list.ItemIndexAtCount(20, 15, 4); got != 1 {
		t.Fatalf("ItemIndexAtCount() first visible row = %d, want 1", got)
	}
	if got := list.ItemIndexAtCount(20, 75, 4); got != -1 {
		t.Fatalf("ItemIndexAtCount() outside visible rows = %d, want -1", got)
	}
}

func TestListViewDefaultContentRectMatchesBounds(t *testing.T) {
	list := NewListView(5, 7, 90, 40, 20, 2, []string{"one"})
	want := list.Rect
	if got := list.ContentRect(); got != want {
		t.Fatalf("default ContentRect() = %+v, want %+v", got, want)
	}
	if got := list.ItemIndexAt(6, 8); got != 0 {
		t.Fatalf("default ItemIndexAt() = %d, want 0", got)
	}
}
