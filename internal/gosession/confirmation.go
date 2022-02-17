package gosession

/*
	CONFIRMATION SYSTEM

	The confirmation system is a way to determine what choices users have made
	when presented with questions posed by the system. Generally but not limited to agreement
	of terms and conditions.

	A notification consists primarily in the form of a question or text presented to the user.
	When they make their decision it is stored in the database and the UI and/or permissions will
	reflect their decision.

*/

// UserConfirmation struct
type UserConfirmation struct {
	Version  int    `json:"version,omitempty" bson:"version,omitempty"`
	Message  string `json:"message,omitempty" bson:"message,omitempty"`
	Accepted bool   `json:"accepted,omitempty" bson:"accepted,omitempty"`
}
