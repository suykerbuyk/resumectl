package prompts

import "github.com/jsuykerbuyk/resumectl/internal/vault"

// DecomposeContext provides data for the decompose_resume template.
type DecomposeContext struct {
	SourceText string
}

// ExtractJobContext provides data for the extract_job template.
type ExtractJobContext struct {
	PostingText string
}

// GapAnalysisContext provides data for the gap_analysis template.
type GapAnalysisContext struct {
	JobContent      string
	ExperienceFiles []*vault.ExperienceFile
	SkillsInventory string
}

// SynthesizeContext provides data for the synthesize_resume and
// synthesize_cover_letter templates.
type SynthesizeContext struct {
	ProfileSummary  string
	JobContent      string
	ExperienceFiles []*vault.ExperienceFile
	SkillsInventory string
	ContactInfo     string
}

// CoachContext provides data for the coach_question template.
type CoachContext struct {
	Skill           string
	Required        bool
	GapType         string
	JobTitle        string
	JobCompany      string
	ExistingContext string
}
