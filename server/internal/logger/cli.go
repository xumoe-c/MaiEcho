package logger

import (
	"github.com/pterm/pterm"
)

// CLI 可视化工具

type ProgressTracker struct {
	bar *pterm.ProgressbarPrinter
}

func NewProgressTracker(total int, title string) *ProgressTracker {
	bar, _ := pterm.DefaultProgressbar.
		WithTotal(total).
		WithTitle(title).
		WithRemoveWhenDone(true).
		Start()

	return &ProgressTracker{bar: bar}
}

func (p *ProgressTracker) Increment() {
	if p.bar != nil {
		p.bar.Increment()
	}
}

func (p *ProgressTracker) UpdateTitle(title string) {
	if p.bar != nil {
		p.bar.UpdateTitle(title)
	}
}

func (p *ProgressTracker) Stop() {
	if p.bar != nil {
		p.bar.Stop()
	}
}

type Spinner struct {
	spinner *pterm.SpinnerPrinter
}

func StartSpinner(text string) *Spinner {
	spinner, _ := pterm.DefaultSpinner.Start(text)
	return &Spinner{spinner: spinner}
}

func (s *Spinner) Success(text string) {
	if s.spinner != nil {
		s.spinner.Success(text)
	}
}

func (s *Spinner) Fail(text string) {
	if s.spinner != nil {
		s.spinner.Fail(text)
	}
}

func (s *Spinner) UpdateText(text string) {
	if s.spinner != nil {
		s.spinner.UpdateText(text)
	}
}

func InfoBox(title, text string) {
	pterm.DefaultBox.WithTitle(title).Println(text)
}

func Section(title string) {
	pterm.DefaultSection.Println(title)
}
