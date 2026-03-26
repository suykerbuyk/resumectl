package vault

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testExperienceFiles() []*ExperienceFile {
	return []*ExperienceFile{
		{
			Frontmatter: ExperienceFrontmatter{
				Role:        "Storage Architect",
				CompanySlug: "seagate",
				Tags:        []string{"storage", "nvme", "architecture"},
				Skills:      []string{"NVMe", "Go", "C++"},
				Domain:      "storage",
			},
			RelPath: "experience/seagate.md",
		},
		{
			Frontmatter: ExperienceFrontmatter{
				Role:        "Storage Engineer",
				CompanySlug: "intel",
				Tags:        []string{"storage", "firmware"},
				Skills:      []string{"C", "NVMe", "SPDK"},
				Domain:      "storage",
			},
			RelPath: "experience/intel.md",
		},
		{
			Frontmatter: ExperienceFrontmatter{
				Role:        "Staff Engineer",
				CompanySlug: "cloudscale",
				Tags:        []string{"cloud", "kubernetes", "go"},
				Skills:      []string{"Go", "Kubernetes", "AWS"},
				Domain:      "cloud",
			},
			RelPath: "experience/cloudscale.md",
		},
	}
}

func TestBuildTagIndex(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)

	assert.Len(t, idx.ByTag["storage"], 2)    // seagate + intel
	assert.Len(t, idx.ByTag["cloud"], 1)      // cloudscale
	assert.Len(t, idx.BySkill["nvme"], 2)     // seagate + intel
	assert.Len(t, idx.BySkill["go"], 2)       // seagate + cloudscale
	assert.Len(t, idx.ByDomain["storage"], 2) // seagate + intel
	assert.Len(t, idx.ByDomain["cloud"], 1)   // cloudscale
}

func TestBuildTagIndex_Empty(t *testing.T) {
	idx := BuildTagIndex(nil)
	assert.Empty(t, idx.ByTag)
	assert.Empty(t, idx.BySkill)
	assert.Empty(t, idx.ByDomain)
}

func TestQuery_Tags(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)

	results := idx.Query([]string{"storage"}, nil)
	assert.Len(t, results, 2) // seagate + intel
}

func TestQuery_Skills(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)

	results := idx.Query(nil, []string{"Kubernetes"})
	assert.Len(t, results, 1)
	assert.Equal(t, "experience/cloudscale.md", results[0].RelPath)
}

func TestQuery_TagsAndSkills(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)

	// storage tag matches seagate + intel, Go skill matches seagate + cloudscale
	// Deduplicated: seagate, intel, cloudscale = 3
	results := idx.Query([]string{"storage"}, []string{"Go"})
	assert.Len(t, results, 3)
}

func TestQuery_NoMatches(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)

	results := idx.Query([]string{"blockchain"}, []string{"Haskell"})
	assert.Empty(t, results)
}

func TestQuery_Deduplication(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)

	// NVMe skill matches seagate and intel; storage tag also matches them
	results := idx.Query([]string{"storage"}, []string{"NVMe"})
	assert.Len(t, results, 2) // no duplicates
}

func TestScore(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)

	job := &JobFile{
		Frontmatter: JobFrontmatter{
			RequiredSkills:  []string{"Go", "Kubernetes", "distributed-systems", "storage"},
			PreferredSkills: []string{"Rust", "NVMe", "eBPF"},
		},
	}

	// cloudscale has Go, Kubernetes, and cloud tag (not storage tag)
	// Skills match: Go, Kubernetes = 2 out of 7
	cloudscaleScore := idx.Score(files[2], job)

	// seagate has NVMe, Go, storage tag, nvme tag, architecture tag
	// Skills/tags match: Go, NVMe, storage = 3 out of 7
	seagateScore := idx.Score(files[0], job)

	assert.Greater(t, seagateScore, 0.0)
	assert.Greater(t, cloudscaleScore, 0.0)
	assert.Less(t, cloudscaleScore, 1.0)
	assert.Less(t, seagateScore, 1.0)
}

func TestScore_NilJob(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)
	assert.Equal(t, 0.0, idx.Score(files[0], nil))
}

func TestScore_NoJobSkills(t *testing.T) {
	files := testExperienceFiles()
	idx := BuildTagIndex(files)
	job := &JobFile{Frontmatter: JobFrontmatter{}}
	assert.Equal(t, 0.0, idx.Score(files[0], job))
}

func TestRebuildIndex(t *testing.T) {
	v := openTestVault(t)
	require.Nil(t, v.Index)

	err := v.RebuildIndex()
	require.NoError(t, err)
	require.NotNil(t, v.Index)
	assert.NotEmpty(t, v.Index.ByTag)
}

func TestEnsureIndex(t *testing.T) {
	v := openTestVault(t)
	require.Nil(t, v.Index)

	err := v.EnsureIndex()
	require.NoError(t, err)
	require.NotNil(t, v.Index)

	// Calling again should be a no-op (same index object)
	idx := v.Index
	err = v.EnsureIndex()
	require.NoError(t, err)
	assert.Same(t, idx, v.Index)
}
