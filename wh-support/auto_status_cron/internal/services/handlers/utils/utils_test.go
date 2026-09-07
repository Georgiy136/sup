package handlerutils

import (
	"slices"
	"testing"
)

func TestGenSeqFromInterval(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		start   int64
		end     int64
		want    []int64
		wantErr bool
	}{
		{
			name:  "ascending range",
			start: 1, end: 5,
			want: []int64{1, 2, 3, 4, 5},
		},
		{
			name:  "single element",
			start: 7, end: 7,
			want: []int64{7},
		},
		{
			name:  "invalid: start > end",
			start: 10, end: 2,
			wantErr: true,
		},
		{
			name:  "with negatives",
			start: -2, end: 2,
			want: []int64{-2, -1, 0, 1, 2},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := GenSeqFromInterval(tc.start, tc.end)

			if tc.wantErr {
				if err == nil {
					t.Fatalf("expected error, got nil (start=%d, end=%d)", tc.start, tc.end)
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !slices.Equal(got, tc.want) {
				t.Fatalf("mismatch:\n  got:  %v\n  want: %v", got, tc.want)
			}
		})
	}
}
