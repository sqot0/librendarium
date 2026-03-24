package librusapi

// ResponseHomeWorks models the homeworks payload.
type ResponseHomeWorks struct {
	HomeWorks []HomeWork `json:"HomeWorks"`
	Resources struct {
		HomeWorksCategories Resource `json:"HomeWorks\\Categories"`
		Parent              Resource `json:".."`
	} `json:"Resources"`
	URL string `json:"Url"`
}

type HomeWork struct {
	ID        uint32   `json:"Id"`
	Content   string   `json:"Content"`
	Date      string   `json:"Date"`
	Category  Category `json:"Category"`
	LessonNo  string   `json:"LessonNo"`
	TimeFrom  string   `json:"TimeFrom"`
	TimeTo    string   `json:"TimeTo"`
	CreatedBy struct {
		ID  uint32 `json:"Id"`
		URL string `json:"Url"`
	} `json:"CreatedBy"`
	Class        *SimpleRef `json:"Class"`
	VirtualClass *SimpleRef `json:"VirtualClass"`
	Subject      *SimpleRef `json:"Subject"`
	Classroom    *Classroom `json:"Classroom"`
	AddDate      string     `json:"AddDate"`
}

type Resource struct {
	URL string `json:"Url"`
}

type Category struct {
	ID  uint32 `json:"Id"`
	URL string `json:"Url"`
}

type SimpleRef struct {
	ID  uint32 `json:"Id"`
	URL string `json:"Url"`
}

type Classroom struct {
	ID     uint32 `json:"Id"`
	Symbol string `json:"Symbol"`
	Name   string `json:"Name"`
	Size   uint32 `json:"Size"`
}

// ResponseSubject models subjects.
type ResponseSubject struct {
	Subject struct {
		ID                uint32 `json:"Id"`
		Name              string `json:"Name"`
		No                uint32 `json:"No"`
		Short             string `json:"Short"`
		IsExtracurricular bool   `json:"IsExtracurricular"`
		IsBlockLesson     bool   `json:"IsBlockLesson"`
	} `json:"Subject"`
}

// ResponseCategory models categories.
type ResponseCategory struct {
	Category struct {
		ID    uint32 `json:"Id"`
		Name  string `json:"Name"`
		Color struct {
			ID  uint32 `json:"Id"`
			URL string `json:"Url"`
		} `json:"Color"`
	} `json:"Category"`
}
