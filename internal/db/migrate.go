package db

func InitialMigrate() {
	Db.MustExec(citySchema)
	Db.MustExec(userSchema)
	Db.MustExec(interestSchema)
	Db.MustExec(userInterestsSchema)
	Db.MustExec(sessionSchema)

	for _, shard := range chatDbs {
		shard.MustExec(messageSchema)
	}
}
