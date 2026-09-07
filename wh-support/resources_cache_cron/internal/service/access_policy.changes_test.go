package service

import (
	"testing"
	"time"

	accesspolicy "gitlab.wildberries.ru/wbwh/support/backend/wh-support.git/resources_cache_cron/internal/models/access_policy"
)

func status(v string) *string { return &v }

func applyToSets(policies []accesspolicy.AccessPolicy) map[string]map[int64]struct{} {
	sets := make(map[string]map[int64]struct{})

	for _, p := range policies {
		id := p.TypeAction + "_" + deref(p.StatusId)

		if p.IsDel {
			if members, ok := sets[id]; ok {
				delete(members, p.GroupId)
				if len(members) == 0 {
					delete(sets, id)
				}
			}
			continue
		}

		if sets[id] == nil {
			sets[id] = make(map[int64]struct{})
		}
		sets[id][p.GroupId] = struct{}{}
	}

	return sets
}

func deref(s *string) string {
	if s == nil {
		return "null"
	}
	return *s
}

func TestSortAccessPoliciesForApply_ChurnPairEndsPopulated(t *testing.T) {
	chDt := time.Date(2026, 8, 5, 13, 58, 5, 225779000, time.UTC)

	policies := []accesspolicy.AccessPolicy{
		{CategoryId: 174, TypeAction: "status_perform_category", StatusId: status("BBB"), GroupId: 55, IsDel: false, ChDtParsed: chDt},
		{CategoryId: 174, TypeAction: "status_approve_category", StatusId: status("CCC"), GroupId: 54, IsDel: false, ChDtParsed: chDt},
		{CategoryId: 174, TypeAction: "status_perform_category", StatusId: status("BBB"), GroupId: 55, IsDel: true, ChDtParsed: chDt},
		{CategoryId: 174, TypeAction: "status_approve_category", StatusId: status("CCC"), GroupId: 54, IsDel: true, ChDtParsed: chDt},
	}

	if got := applyToSets(policies); len(got) != 0 {
		t.Fatalf("precondition: expected empty result for ADD->DEL order, got %v", got)
	}

	sortAccessPoliciesForApply(policies)

	got := applyToSets(policies)
	if _, ok := got["status_perform_category_BBB"][55]; !ok {
		t.Errorf("expected group 55 on ...perform_BBB after sort, got %v", got)
	}
	if _, ok := got["status_approve_category_CCC"][54]; !ok {
		t.Errorf("expected group 54 on ...approve_CCC after sort, got %v", got)
	}
}

func TestSortAccessPoliciesForApply_GroupSwap(t *testing.T) {
	chDt := time.Date(2026, 8, 5, 14, 2, 18, 620416000, time.UTC)

	policies := []accesspolicy.AccessPolicy{
		{CategoryId: 174, TypeAction: "status_approve_category", StatusId: status("CCC"), GroupId: 55, IsDel: false, ChDtParsed: chDt},
		{CategoryId: 174, TypeAction: "status_approve_category", StatusId: status("CCC"), GroupId: 54, IsDel: true, ChDtParsed: chDt},
	}

	sortAccessPoliciesForApply(policies)

	got := applyToSets(policies)
	members := got["status_approve_category_CCC"]
	if _, ok := members[55]; !ok {
		t.Errorf("expected new group 55, got %v", members)
	}
	if _, ok := members[54]; ok {
		t.Errorf("old group 54 must not remain, got %v", members)
	}
}

func TestSortAccessPoliciesForApply_ChronologyPreserved(t *testing.T) {
	earlier := time.Date(2026, 8, 5, 10, 0, 0, 0, time.UTC)
	later := earlier.Add(time.Minute)

	policies := []accesspolicy.AccessPolicy{
		{GroupId: 1, IsDel: false, ChDtParsed: later},
		{GroupId: 1, IsDel: true, ChDtParsed: earlier},
	}

	sortAccessPoliciesForApply(policies)

	if !policies[0].ChDtParsed.Equal(earlier) {
		t.Errorf("expected earlier change first, got %v", policies[0].ChDtParsed)
	}
}
