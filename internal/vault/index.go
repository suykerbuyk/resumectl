package vault

import (
	"strings"
)

// TagIndex provides fast lookups of experience files by tag, skill, and domain.
type TagIndex struct {
	ByTag    map[string][]*ExperienceFile
	BySkill  map[string][]*ExperienceFile
	ByDomain map[string][]*ExperienceFile
}

// BuildTagIndex creates a TagIndex from a slice of experience files.
func BuildTagIndex(files []*ExperienceFile) *TagIndex {
	idx := &TagIndex{
		ByTag:    make(map[string][]*ExperienceFile),
		BySkill:  make(map[string][]*ExperienceFile),
		ByDomain: make(map[string][]*ExperienceFile),
	}

	for _, f := range files {
		for _, tag := range f.Frontmatter.Tags {
			key := strings.ToLower(tag)
			idx.ByTag[key] = append(idx.ByTag[key], f)
		}
		for _, skill := range f.Frontmatter.Skills {
			key := strings.ToLower(skill)
			idx.BySkill[key] = append(idx.BySkill[key], f)
		}
		if f.Frontmatter.Domain != "" {
			key := strings.ToLower(f.Frontmatter.Domain)
			idx.ByDomain[key] = append(idx.ByDomain[key], f)
		}
	}

	return idx
}

// Query returns experience files matching any of the given tags or skills.
// Results are deduplicated — each file appears at most once.
func (idx *TagIndex) Query(tags []string, skills []string) []*ExperienceFile {
	seen := make(map[string]bool)
	var results []*ExperienceFile

	add := func(files []*ExperienceFile) {
		for _, f := range files {
			if !seen[f.RelPath] {
				seen[f.RelPath] = true
				results = append(results, f)
			}
		}
	}

	for _, tag := range tags {
		add(idx.ByTag[strings.ToLower(tag)])
	}
	for _, skill := range skills {
		add(idx.BySkill[strings.ToLower(skill)])
	}

	return results
}

// Score computes a relevance score (0.0–1.0) for an experience file against a job.
// The score is the fraction of the job's required + preferred skills that match
// the experience file's skills or tags.
func (idx *TagIndex) Score(file *ExperienceFile, job *JobFile) float64 {
	if job == nil {
		return 0
	}

	fileKeys := make(map[string]bool)
	for _, s := range file.Frontmatter.Skills {
		fileKeys[strings.ToLower(s)] = true
	}
	for _, t := range file.Frontmatter.Tags {
		fileKeys[strings.ToLower(t)] = true
	}

	allJobSkills := make([]string, 0, len(job.Frontmatter.RequiredSkills)+len(job.Frontmatter.PreferredSkills))
	allJobSkills = append(allJobSkills, job.Frontmatter.RequiredSkills...)
	allJobSkills = append(allJobSkills, job.Frontmatter.PreferredSkills...)

	if len(allJobSkills) == 0 {
		return 0
	}

	matches := 0
	for _, s := range allJobSkills {
		if fileKeys[strings.ToLower(s)] {
			matches++
		}
	}

	return float64(matches) / float64(len(allJobSkills))
}
