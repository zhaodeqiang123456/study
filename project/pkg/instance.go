package pkg

type Instance struct {
	dbS *dbService
}

func (inst *Instance) InitInstance() {
	inst.dbS = &dbService{}
	inst.dbS.initDB()
}

func (inst *Instance) Detach() {
	inst.dbS.db.Close()
}

func (inst *Instance) GetDbS() *dbService {
	return inst.dbS
}
