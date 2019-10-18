package db

func InitialMigrate() {
	Db.MustExec("CREATE DATABASE IF NOT EXISTS social")
	Db.MustExec(citySchema)
	Db.MustExec(userSchema)
	Db.MustExec(interestSchema)
	Db.MustExec(userInterestsSchema)
	Db.MustExec(sessionSchema)
}
