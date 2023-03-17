package mongodb

import "go.mongodb.org/mongo-driver/mongo"

const (
	errorCodeDuplicateKey = 11000
)

func isMongoError(err error, code int) bool {
	merr, ok := err.(mongo.WriteException) // TODO(mredolatti): handle other errors as well
	if !ok {
		return false
	}

	for idx := range merr.WriteErrors {
		if merr.WriteErrors[idx].Code == code {
			return true
		}
	}
	return false
}
