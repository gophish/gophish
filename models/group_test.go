package models

import (
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"gopkg.in/check.v1"
)

func (s *ModelsSuite) TestPostGroup(c *check.C) {
	g := Group{Name: "Test Group"}
	g.Targets = []Target{Target{BaseRecipient: BaseRecipient{Email: "test@example.com"}}}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, nil)
	c.Assert(g.Name, check.Equals, "Test Group")
	c.Assert(g.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestPostGroupNoName(c *check.C) {
	g := Group{Name: ""}
	g.Targets = []Target{Target{BaseRecipient: BaseRecipient{Email: "test@example.com"}}}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, ErrGroupNameNotSpecified)
}

func (s *ModelsSuite) TestPostGroupNoTargets(c *check.C) {
	g := Group{Name: "No Target Group"}
	g.Targets = []Target{}
	g.UserId = 1
	err := PostGroup(&g)
	c.Assert(err, check.Equals, ErrNoTargetsSpecified)
}

func (s *ModelsSuite) TestGetGroups(c *check.C) {
	// Add groups.
	PostGroup(&Group{
		Name: "Test Group 1",
		Targets: []Target{
			Target{
				BaseRecipient: BaseRecipient{Email: "test1@example.com"},
			},
		},
		UserId: 1,
	})
	PostGroup(&Group{
		Name: "Test Group 2",
		Targets: []Target{
			Target{
				BaseRecipient: BaseRecipient{Email: "test2@example.com"},
			},
		},
		UserId: 1,
	})

	// Get groups and test result.
	groups, err := GetGroups(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(groups), check.Equals, 2)
	c.Assert(len(groups[0].Targets), check.Equals, 1)
	c.Assert(len(groups[1].Targets), check.Equals, 1)
	c.Assert(groups[0].Name, check.Equals, "Test Group 1")
	c.Assert(groups[1].Name, check.Equals, "Test Group 2")
	c.Assert(groups[0].Targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(groups[1].Targets[0].Email, check.Equals, "test2@example.com")
}

func (s *ModelsSuite) TestGetGroupsNoGroups(c *check.C) {
	groups, err := GetGroups(1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(groups), check.Equals, 0)
}

func (s *ModelsSuite) TestGetGroup(c *check.C) {
	// Add group.
	originalGroup := &Group{
		Name: "Test Group",
		Targets: []Target{
			Target{
				BaseRecipient: BaseRecipient{Email: "test@example.com"},
			},
		},
		UserId: 1,
	}
	c.Assert(PostGroup(originalGroup), check.Equals, nil)

	// Get group and test result.
	group, err := GetGroup(originalGroup.Id, 1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(group.Targets), check.Equals, 1)
	c.Assert(group.Name, check.Equals, "Test Group")
	c.Assert(group.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestGetGroupNoGroups(c *check.C) {
	_, err := GetGroup(1, 1)
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestGetGroupByName(c *check.C) {
	// Add group.
	PostGroup(&Group{
		Name: "Test Group",
		Targets: []Target{
			Target{
				BaseRecipient: BaseRecipient{Email: "test@example.com"},
			},
		},
		UserId: 1,
	})

	// Get group and test result.
	group, err := GetGroupByName("Test Group", 1)
	c.Assert(err, check.Equals, nil)
	c.Assert(len(group.Targets), check.Equals, 1)
	c.Assert(group.Name, check.Equals, "Test Group")
	c.Assert(group.Targets[0].Email, check.Equals, "test@example.com")
}

func (s *ModelsSuite) TestGetGroupByNameNoGroups(c *check.C) {
	_, err := GetGroupByName("Test Group", 1)
	c.Assert(err, check.Equals, gorm.ErrRecordNotFound)
}

func (s *ModelsSuite) TestPutGroup(c *check.C) {
	// Add test group.
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{BaseRecipient: BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		Target{BaseRecipient: BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	PostGroup(&group)

	// Update one of group's targets.
	group.Targets[0].FirstName = "Updated"
	err := PutGroup(&group)
	c.Assert(err, check.Equals, nil)

	// Verify updated target information.
	targets, _ := GetTargets(group.Id)
	c.Assert(targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(targets[0].FirstName, check.Equals, "Updated")
	c.Assert(targets[0].LastName, check.Equals, "Example")
	c.Assert(targets[1].Email, check.Equals, "test2@example.com")
	c.Assert(targets[1].FirstName, check.Equals, "Second")
	c.Assert(targets[1].LastName, check.Equals, "Example")
}

func (s *ModelsSuite) TestPutGroupEmptyAttribute(c *check.C) {
	// Add test group.
	group := Group{Name: "Test Group"}
	group.Targets = []Target{
		Target{BaseRecipient: BaseRecipient{Email: "test1@example.com", FirstName: "First", LastName: "Example"}},
		Target{BaseRecipient: BaseRecipient{Email: "test2@example.com", FirstName: "Second", LastName: "Example"}},
	}
	group.UserId = 1
	PostGroup(&group)

	// Update one of group's targets.
	group.Targets[0].FirstName = ""
	err := PutGroup(&group)
	c.Assert(err, check.Equals, nil)

	// Verify updated empty attribute was saved.
	targets, _ := GetTargets(group.Id)
	c.Assert(targets[0].Email, check.Equals, "test1@example.com")
	c.Assert(targets[0].FirstName, check.Equals, "")
	c.Assert(targets[0].LastName, check.Equals, "Example")
	c.Assert(targets[1].Email, check.Equals, "test2@example.com")
	c.Assert(targets[1].FirstName, check.Equals, "Second")
	c.Assert(targets[1].LastName, check.Equals, "Example")
}

func benchmarkPostGroup(b *testing.B, iter, size int) {
	b.StopTimer()
	g := &Group{
		Name: fmt.Sprintf("Group-%d", iter),
	}
	for i := 0; i < size; i++ {
		g.Targets = append(g.Targets, Target{
			BaseRecipient: BaseRecipient{
				FirstName: "User",
				LastName:  fmt.Sprintf("%d", i),
				Email:     fmt.Sprintf("test-%d@test.com", i),
			},
		})
	}
	b.StartTimer()
	err := PostGroup(g)
	if err != nil {
		b.Fatalf("error posting group: %v", err)
	}
}

// benchmarkPutGroup modifies half of the group to simulate a large change
func benchmarkPutGroup(b *testing.B, iter, size int) {
	b.StopTimer()
	// First, we need to create the group
	g := &Group{
		Name: fmt.Sprintf("Group-%d", iter),
	}
	for i := 0; i < size; i++ {
		g.Targets = append(g.Targets, Target{
			BaseRecipient: BaseRecipient{
				FirstName: "User",
				LastName:  fmt.Sprintf("%d", i),
				Email:     fmt.Sprintf("test-%d@test.com", i),
			},
		})
	}
	err := PostGroup(g)
	if err != nil {
		b.Fatalf("error posting group: %v", err)
	}
	// Now we need to change half of the group.
	for i := 0; i < size/2; i++ {
		g.Targets[i].Email = fmt.Sprintf("test-modified-%d@test.com", i)
	}
	b.StartTimer()
	err = PutGroup(g)
	if err != nil {
		b.Fatalf("error modifying group: %v", err)
	}
}

func BenchmarkPostGroup100(b *testing.B) {
	setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkPostGroup(b, i, 100)
		b.StopTimer()
		resetBenchmark(b)
	}
	tearDownBenchmark(b)
}

func BenchmarkPostGroup1000(b *testing.B) {
	setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkPostGroup(b, i, 1000)
		b.StopTimer()
		resetBenchmark(b)
	}
	tearDownBenchmark(b)
}

func BenchmarkPostGroup10000(b *testing.B) {
	setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkPostGroup(b, i, 10000)
		b.StopTimer()
		resetBenchmark(b)
	}
	tearDownBenchmark(b)
}

func BenchmarkPutGroup100(b *testing.B) {
	setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkPutGroup(b, i, 100)
		b.StopTimer()
		resetBenchmark(b)
	}
	tearDownBenchmark(b)
}

func BenchmarkPutGroup1000(b *testing.B) {
	setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkPutGroup(b, i, 1000)
		b.StopTimer()
		resetBenchmark(b)
	}
	tearDownBenchmark(b)
}

func BenchmarkPutGroup10000(b *testing.B) {
	setupBenchmark(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		benchmarkPutGroup(b, i, 10000)
		b.StopTimer()
		resetBenchmark(b)
	}
	tearDownBenchmark(b)
}
