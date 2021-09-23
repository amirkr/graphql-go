package model

type Author struct {
	ID        string
	Firstname string
	Lastname  string
	Createdat string
	Object struct {
		Obj_id int
		Obj_name string
	}
}