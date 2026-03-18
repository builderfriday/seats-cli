package model

// Program represents a loyalty program with its identifier and display name.
type Program struct {
	ID   string
	Name string
}

// Programs is the list of supported loyalty programs.
var Programs = []Program{
	{ID: "aeroplan", Name: "Air Canada Aeroplan"},
	{ID: "alaska", Name: "Alaska Mileage Plan"},
	{ID: "american", Name: "American AAdvantage"},
	{ID: "delta", Name: "Delta SkyMiles"},
	{ID: "etihad", Name: "Etihad Guest"},
	{ID: "eurobonus", Name: "SAS EuroBonus"},
	{ID: "flyingblue", Name: "Air France/KLM Flying Blue"},
	{ID: "lifemiles", Name: "Avianca LifeMiles"},
	{ID: "qantas", Name: "Qantas Frequent Flyer"},
	{ID: "united", Name: "United MileagePlus"},
	{ID: "velocity", Name: "Virgin Australia Velocity"},
	{ID: "virginatlantic", Name: "Virgin Atlantic Flying Club"},
}
