package main

type instance struct {
	dbS *dbService
}

func (inst *instance) initInstance() {
	inst.dbS = &dbService{}
	inst.dbS.initDB()
}

func (inst *instance) detach() {
	inst.dbS.db.Close()
}
