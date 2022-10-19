package models

import "github.com/uptrace/bun"

type BasicBlock struct {
	bun.BaseModel

	ID            int64 `bun:",pk,autoincrement"`
	Address       uint32
	LastAddress   uint32
	BranchAddress uint32
}
