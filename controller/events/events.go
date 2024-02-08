package events

// COSI relevant event reasons
const (
	FailedCreateBucket = "FailedCreateBucket"
	FailedDeleteBucket = "FailedDeleteBucket"
	WaitingForBucket   = "WaitingForBucket"

	FailedGrantAccess  = "FailedGrantAccess"
	FailedRevokeAccess = "FailedRevokeAccess"
)
