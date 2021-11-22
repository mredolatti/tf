package main

type Medico struct {
	id string
}

func (m *Medico) ID() string {
	return m.id
}

type Institucion struct {
	id string
}

func (i *Institucion) ID() string {
	return i.id
}

type Paciente struct {
	id string
}

func (p *Paciente) ID() string {
	return p.id
}

type Archivo struct {
	id            string
	idPaciente    string
	idInstitucion string
	ruta          string
}

func (a *Archivo) ID() string {
	return a.id
}

func (a *Archivo) IDPaciente() string {
	return a.idPaciente
}

func (a *Archivo) Ruta() string {
	return a.ruta
}

type ACL struct {
	rutaArchivo string
	idMedico    string
	permisos    uint
}
