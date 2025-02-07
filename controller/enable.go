package controller

import (
	"bytes"
	"encoding/json"
	"github.com/go-openapi/runtime/middleware"
	"github.com/jmoiron/sqlx"
	"github.com/openziti/zrok/controller/store"
	"github.com/openziti/zrok/controller/zrokEdgeSdk"
	"github.com/openziti/zrok/rest_model_zrok"
	"github.com/openziti/zrok/rest_server_zrok/operations/environment"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
)

type enableHandler struct {
	cfg *LimitsConfig
}

func newEnableHandler(cfg *LimitsConfig) *enableHandler {
	return &enableHandler{cfg: cfg}
}

func (h *enableHandler) Handle(params environment.EnableParams, principal *rest_model_zrok.Principal) middleware.Responder {
	// start transaction early; if it fails, don't bother creating ziti resources
	tx, err := str.Begin()
	if err != nil {
		logrus.Errorf("error starting transaction for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}
	defer func() { _ = tx.Rollback() }()

	if err := h.checkLimits(principal, tx); err != nil {
		logrus.Errorf("limits error for user '%v': %v", principal.Email, err)
		return environment.NewEnableUnauthorized()
	}

	client, err := edgeClient()
	if err != nil {
		logrus.Errorf("error getting edge client for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	uniqueToken, err := createShareToken()
	if err != nil {
		logrus.Errorf("error creating unique identity token for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	ident, err := zrokEdgeSdk.CreateEnvironmentIdentity(uniqueToken, principal.Email, params.Body.Description, client)
	if err != nil {
		logrus.Errorf("error creating environment identity for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	envZId := ident.Payload.Data.ID
	cfg, err := zrokEdgeSdk.EnrollIdentity(envZId, client)
	if err != nil {
		logrus.Errorf("error enrolling environment identity for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	if err := zrokEdgeSdk.CreateEdgeRouterPolicy(envZId, envZId, client); err != nil {
		logrus.Errorf("error creating edge router policy for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}

	envId, err := str.CreateEnvironment(int(principal.ID), &store.Environment{
		Description: params.Body.Description,
		Host:        params.Body.Host,
		Address:     realRemoteAddress(params.HTTPRequest),
		ZId:         envZId,
	}, tx)
	if err != nil {
		logrus.Errorf("error storing created identity for user '%v': %v", principal.Email, err)
		_ = tx.Rollback()
		return environment.NewEnableInternalServerError()
	}

	if err := tx.Commit(); err != nil {
		logrus.Errorf("error committing for user '%v': %v", principal.Email, err)
		return environment.NewEnableInternalServerError()
	}
	logrus.Infof("created environment for '%v', with ziti identity '%v', and database id '%v'", principal.Email, ident.Payload.Data.ID, envId)

	resp := environment.NewEnableCreated().WithPayload(&rest_model_zrok.EnableResponse{
		Identity: envZId,
	})

	var out bytes.Buffer
	enc := json.NewEncoder(&out)
	enc.SetEscapeHTML(false)
	err = enc.Encode(&cfg)
	if err != nil {
		panic(err)
	}
	resp.Payload.Cfg = out.String()

	return resp
}

func (h *enableHandler) checkLimits(principal *rest_model_zrok.Principal, tx *sqlx.Tx) error {
	if !principal.Limitless && h.cfg.Environments > Unlimited {
		envs, err := str.FindEnvironmentsForAccount(int(principal.ID), tx)
		if err != nil {
			return errors.Errorf("unable to find environments for account '%v': %v", principal.Email, err)
		}
		if len(envs)+1 > h.cfg.Environments {
			return errors.Errorf("would exceed environments limit of %d for '%v'", h.cfg.Environments, principal.Email)
		}
	}
	return nil
}
