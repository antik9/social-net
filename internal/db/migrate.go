package db

func InitialMigrate() {
	Db.MustExec(citySchema)
	Db.MustExec(userSchema)
	Db.MustExec(interestSchema)
	Db.MustExec(userInterestsSchema)
	Db.MustExec(sessionSchema)
	Db.MustExec(feedMessageSchema)
	Db.MustExec(subscriptionSchema)

	for _, shard := range chatDbs {
		shard.MustExec(messageSchema)
		shard.MustExec(feedUserLinkSchema)
	}
}
