package main

// DO NOT EDIT!
// This code is generated with http://github.com/hexdigest/gowrap tool
// using https://raw.githubusercontent.com/hexdigest/gowrap/bd05dcaf6963696b62ac150a98a59674456c6c53/templates/log template

//go:generate gowrap gen -d . -i Registrater -t https://raw.githubusercontent.com/hexdigest/gowrap/bd05dcaf6963696b62ac150a98a59674456c6c53/templates/log -o registrater_with_logging.go

import (
	"context"
	"io"
	"log"
)

// RegistraterWithLog implements Registrater that is instrumented with logging
type RegistraterWithLog struct {
	_stdlog, _errlog *log.Logger
	_base            Registrater
}

// NewRegistraterWithLog instruments an implementation of the Registrater with simple logging
func NewRegistraterWithLog(base Registrater, stdout, stderr io.Writer) RegistraterWithLog {
	return RegistraterWithLog{
		_base:   base,
		_stdlog: log.New(stdout, "", log.LstdFlags),
		_errlog: log.New(stderr, "", log.LstdFlags),
	}
}

// Registrate implements Registrater
func (_d RegistraterWithLog) Registrate(ctx context.Context, fp1 *Form) (up1 *User, err error) {
	_params := []interface{}{"RegistraterWithLog: calling Registrate with params:", ctx, fp1}
	_d._stdlog.Println(_params...)
	defer func() {
		_results := []interface{}{"RegistraterWithLog: Registrate returned results:", up1, err}
		if err != nil {
			_d._errlog.Println(_results...)
		} else {
			_d._stdlog.Println(_results...)
		}
	}()
	return _d._base.Registrate(ctx, fp1)
}
