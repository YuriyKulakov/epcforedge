// SPDX-License-Identifier: Apache-2.0
// Copyright © 2020 Intel Corporation

package af

import (
	"context"
	"encoding/json"
	"net/http"
)

func getAllPfdTransactions(cliCtx context.Context, afCtx *Context) (
	[]PfdManagement, *http.Response, error) {

	cliCfg := NewConfiguration(afCtx)
	cli := NewClient(cliCfg)

	tTrans, resp, err := cli.PfdManagementGetAllAPI.PfdTransactionsGetAll(
		cliCtx, afCtx.cfg.AfID)

	if err != nil {
		return nil, resp, err
	}
	return tTrans, resp, nil

}

// GetAllPfdTransactions - Function to read all PFD transactions
func GetAllPfdTransactions(w http.ResponseWriter, r *http.Request) {
	var (
		err        error
		tsResp     []PfdManagement
		resp       *http.Response
		tsRespJSON []byte
	)

	afCtx := r.Context().Value(keyType("af-ctx")).(*Context)
	if afCtx == nil {
		log.Errf("Pfd Management get all: " +
			"af-ctx retrieved from request is nil")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cliCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	tsResp, resp, err = getAllPfdTransactions(cliCtx, afCtx)
	if err != nil {
		log.Errf("PFD Management transactions get all : %s", err.Error())
		w.WriteHeader(getStatusCode(resp))
		return
	}

	for key, v := range tsResp {
		// Updating the Self Link and Applications Self Link in AF
		var self string
		self, err = updateSelfLink(afCtx.cfg, r, v)
		if err != nil {
			log.Errf("PFD Management transactions get all : %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		v.Self = Link(self)
		err = updateAppsLink(afCtx.cfg, r, v)
		if err != nil {
			log.Errf("PFD Management transactions get all : %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tsResp[key] = v
	}

	tsRespJSON, err = json.Marshal(tsResp)
	if err != nil {
		log.Errf("PFD Management transactions get all : %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(resp.StatusCode)

	if _, err = w.Write(tsRespJSON); err != nil {
		log.Errf("PFD Management transactions get all : %s", err.Error())
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}
