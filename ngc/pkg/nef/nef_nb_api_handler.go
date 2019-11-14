// Copyright 2019 Intel Corporation and Smart-Edge.com, Inc. All rights reserved
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

const SUB_START_ID int = 11111
const MAX_SUB_SUPP int = 5

//AF Subscription data
type afSubscription struct {
	subid string
	loc   string
	ti    TrafficInfluSub
}

//AF data
type afData struct {
	afid       string
	subIdnum   int
	maxSubSupp int
	subs       map[string]*afSubscription
	//Member functions
	//afCreate
	//afAddSubscription
	//afGetSubscription
	//afUpdateSubscription
	//afDeleteSubscription
	//afDestroy
}

//Creates a AF instance
func (af *afData) afCreate(afid string) error {

	//Validate afid ??

	af.afid = afid
	af.subIdnum = SUB_START_ID //Start Number
	af.maxSubSupp = MAX_SUB_SUPP
	af.subs = make(map[string]*afSubscription)
	return nil
}

//Creates a new subscription
func (af *afData) afAddSubscription(ti TrafficInfluSub) (loc string, err error) {

	if af.subIdnum >= nef.maxSubSup+SUB_START_ID {
		return "", errors.New("MAX SUBS Created")
	}
	log.Println(af.subIdnum)
	log.Println(af.maxSubSupp)
	//Generate a unique subscription ID string
	subIdStr := strconv.Itoa(af.subIdnum)
	af.subIdnum++
	log.Println(af.subIdnum)

	//Create the Subscription info
	afsub := afSubscription{subid: subIdStr, ti: ti}

	//Create Location URI
	afsub.loc = nef.location + af.afid + "/subscriptions/" + subIdStr

	//Link the subscription with the AF
	af.subs[subIdStr] = &afsub

	log.Println(" SUBSCRIPTION ADDED ")
	log.Println(len(af.subs))
	//log.Println(af.subs)

	return afsub.loc, nil
}

func (af *afData) afUpdateSubscription(subId string, ti TrafficInfluSub) (err error) {

	sub, ok := af.subs[subId]
	if ok == false {
		sub.ti = ti
		return errors.New("Subscription Not Found")
	}
	sub.ti = ti
	return
}

func (af *afData) afGetSubscription(subId string) (ti TrafficInfluSub, err error) {

	_, ok := af.subs[subId]

	if ok == true {
		return af.subs[subId].ti, nil
	}

	return ti, errors.New("SubscriptionId Not found")

}

func (af *afData) afGetSubscriptionList() (subslist []TrafficInfluSub, err error) {

	if len(af.subs) > 0 {
		for _, value := range af.subs {
			subslist = append(subslist, value.ti)
		}
		return subslist, nil
	}

	return nil, errors.New("No Subscriptions present")
}

func (af *afData) afDeleteSubscription(subId string) error {
	//Check if AF is already present
	_, ok := af.subs[subId]

	if ok == true {
		delete(af.subs, subId)
		af.subIdnum--
		return nil
	} else {
		return errors.New("SubscriotionId not found")
	}
}
func (af *afData) afDestroy(afid string) error {

	//Todo delete all subscriptions, needed in go ??
	//Needed for gracefully disconnecting
	return errors.New("AF data cleaned")
}

//NEF context data
type nefData struct {
	nefport   string
	location  string
	afcount   int
	subIdnum  int
	maxSubSup int
	maxAfSup  int
	afs       map[string]*afData

	//Member functions
	//nefCreate
	//nefAddAf
	//GetAf
	//DeleteAf
	//Destroy
}

//NEF context info data
var nef nefData

//Initialize the NEF component
func (nef *nefData) nefCreate() error {

	//To be fetched from config right now hard coded
	nef.nefport = "80"
	nef.location = "http://localhost:80/3gpp-traffic-influence/v1/"
	nef.afcount = 0
	nef.subIdnum = 11111
	nef.maxAfSup = 2
	nef.maxSubSup = 5
	nef.afs = make(map[string]*afData)

	return nil
}

func NEFInit() error {

	return nef.nefCreate()
}

func (nef *nefData) nefAddAf(afId string) (af *afData, err error) {

	var afe afData

	//Check if AF is already present
	_, ok := nef.afs[afId]

	if ok == true {
		return nef.afs[afId], errors.New("AF already present")
	} else {
		//Create a new entry of AF

		afe.afCreate(afId)
		nef.afs[afId] = &afe
		nef.afcount++
	}
	return &afe, nil
}

func (nef *nefData) nefGetAf(afId string) (af *afData, err error) {

	//Check if AF is already present
	afe, ok := nef.afs[afId]

	if ok == true {
		return afe, nil
	} else {
		err = errors.New("AF entry not present")
		return afe, err
	}
}

func (nef *nefData) nefDeleteAf(afId string) (err error) {

	//Check if AF is already present
	_, ok := nef.afs[afId]

	if ok == true {
		delete(nef.afs, afId)
		nef.afcount--
		return nil
	} else {
		err = errors.New("AF entry not present")
		return err
	}
}

func (nef *nefData) nefDestroy() {

	// Todo
}

func createNewSub(afId string, ti TrafficInfluSub) (loc string, err error) {

	var af *afData

	//Validate the Traffic Influence
	err = validateTIS(ti)
	if err != nil {
		log.Println(err)
		return "", err
	}

	af, err = nef.nefGetAf(afId)

	if err != nil {
		log.Println("NO AF PRESENT CREATE AF")
		af, _ = nef.nefAddAf(afId)
	} else {
		log.Println("AF PRESENT AF")
		log.Println(af)
	}

	loc, err = af.afAddSubscription(ti)

	if err != nil {
		log.Println(err)
		return loc, err
	}

	//log.Println(nef)
	//log.Printf("AF COUNT: %+v", nef.afcount)

	return loc, nil
}

//Validate the Traffic influence data received from AF
func validateTIS(ti TrafficInfluSub) (err error) {

	//Check if we have crossed max supported AF
	if nef.afcount >= nef.maxAfSup {
		log.Println("MAX AF exceeded ")
		return errors.New("MAX AF exceeded")
		//return err
	}
	return nil
}

// ReadAllTrafficInfluenceSubscription : API to read all the subscritions
func ReadAllTrafficInfluenceSubscription(w http.ResponseWriter,
	r *http.Request) {

	log.Printf("===============================================")
	log.Printf(" Method : GET ")
	log.Printf(" URL PATH : " + r.URL.Path[1:])

	vars := mux.Vars(r)

	log.Printf(" AFID  : %s", vars["afId"])

	af, _ := nef.nefGetAf(vars["afId"])

	subslist, _ := af.afGetSubscriptionList()

	mdata, err2 := json.Marshal(subslist)
	if err2 != nil {
		log.Printf("Error:  Failed to marshal the json data")
		log.Println(err2)
		//panic(err2)
		w.WriteHeader(http.StatusOK)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(mdata)

	w.WriteHeader(http.StatusOK)
}

// CreateTrafficInfluenceSubscription : Handles the traffic influence requested
// by AF
func CreateTrafficInfluenceSubscription(w http.ResponseWriter,
	r *http.Request) {

	log.Printf("===============================================")
	log.Printf(" Method : POST ")
	log.Printf(" URL PATH : " + r.URL.Path[1:])
	vars := mux.Vars(r)
	log.Printf(" AFID  : %s", vars["afId"])

	b, err := ioutil.ReadAll(r.Body)

	defer r.Body.Close()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	//Traffic Influence data
	TrInBody := TrafficInfluSub{}

	//Convert the json Traffic Influence data into struct
	err1 := json.Unmarshal(b, &TrInBody)

	//Print
	//log.Printf("\n Traffic Influence data from AF\n%+v\n\n", TrInBody)

	if err1 != nil {
		log.Printf("Error: Failed to UNmarshal POST req ")
		log.Println(err1)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	//Update the data respose of POST

	//Martshal data and send into the body
	mdata, err2 := json.Marshal(TrInBody)

	if err2 != nil {
		log.Printf("Error:  Failed to marshal the json data")
		log.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
	}

	//loc, err3 := createNewSubscription(vars["afId"], TrInBody)
	loc, err3 := createNewSub(vars["afId"], TrInBody)
	log.Println(loc)

	logNef()

	if err3 != nil {
		log.Printf("Error:  Failed to Create AF data")
		log.Println(err3)
		//panic(err3)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Header().Set("Location", loc)

	w.Write(mdata)

	w.WriteHeader(http.StatusOK)
}

// ReadTrafficInfluenceSubscription : Read a particular subscription details
func ReadTrafficInfluenceSubscription(w http.ResponseWriter, r *http.Request) {

	log.Printf("===============================================")
	log.Printf(" Method : GET ")
	log.Printf(" URL PATH : " + r.URL.Path[1:])
	vars := mux.Vars(r)
	log.Printf(" AFID  : %s", vars["afId"])
	log.Printf(" SUBSCRIPTION ID  : %s", vars["subscriptionId"])

	af, ok := nef.nefGetAf(vars["afId"])

	if ok != nil {
		log.Println(ok)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Println("AF Found ")

	substi, err := af.afGetSubscription(vars["subscriptionId"])

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	mdata, err2 := json.Marshal(substi)
	if err2 != nil {
		log.Printf("Error:  Failed to marshal the json data")
		log.Println(err2)
		w.WriteHeader(http.StatusInternalServerError)
	}
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.Write(mdata)

	w.WriteHeader(http.StatusOK)
}

// UpdatePutTrafficInfluenceSubscription : Updates a traffic influence created
// earlier (PUT Req)
func UpdatePutTrafficInfluenceSubscription(w http.ResponseWriter,
	r *http.Request) {

	log.Printf("===============================================")
	log.Printf(" Method : PUT ")
	log.Printf(" URL PATH : " + r.URL.Path[1:])

	vars := mux.Vars(r)

	log.Printf(" AFID  : %s", vars["afId"])
	log.Printf(" SUBSCRIPTION ID  : %s", vars["subscriptionId"])

	af, ok := nef.nefGetAf(vars["afId"])
	if ok == nil {

		b, err := ioutil.ReadAll(r.Body)

		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Traffic Influence data
		TrInBody := TrafficInfluSub{}

		//Convert the json Traffic Influence data into struct
		err1 := json.Unmarshal(b, &TrInBody)

		if err1 != nil {
			log.Printf("Error: Failed to UNmarshal POST req ")
			log.Println(err1)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ok := af.afUpdateSubscription(vars["subscriptionId"], TrInBody)
		if ok != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		mdata, err2 := json.Marshal(TrInBody)

		if err2 != nil {
			log.Printf("Error:  Failed to marshal the json data")
			log.Println(err2)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		w.Write(mdata)

	}

	w.WriteHeader(http.StatusOK)
}

// UpdatePatchTrafficInfluenceSubscription : Updates a traffic influence created
//  earlier (PATCH Req)
func UpdatePatchTrafficInfluenceSubscription(w http.ResponseWriter,
	r *http.Request) {

	log.Printf("===============================================")
	log.Printf(" Method : PATCH ")
	log.Printf(" URL PATH : " + r.URL.Path[1:])

	vars := mux.Vars(r)

	log.Printf(" AFID  : %s", vars["afId"])
	log.Printf(" SUBSCRIPTION ID  : %s", vars["subscriptionId"])

	af, ok := nef.nefGetAf(vars["afId"])
	if ok == nil {

		b, err := ioutil.ReadAll(r.Body)

		defer r.Body.Close()

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		//Traffic Influence data
		TrInBody := TrafficInfluSub{}

		//Convert the json Traffic Influence data into struct
		err1 := json.Unmarshal(b, &TrInBody)

		if err1 != nil {
			log.Printf("Error: Failed to UNmarshal POST req ")
			log.Println(err1)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		ok := af.afUpdateSubscription(vars["subscriptionId"], TrInBody)
		if ok != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		mdata, err2 := json.Marshal(TrInBody)

		if err2 != nil {
			log.Printf("Error:  Failed to marshal the json data")
			log.Println(err2)
			w.WriteHeader(http.StatusInternalServerError)
		}
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")

		w.Write(mdata)

	}

	w.WriteHeader(http.StatusOK)
}

// DeleteTrafficInfluenceSubscription : Deletes a traffic influence created by
//  AF
func DeleteTrafficInfluenceSubscription(w http.ResponseWriter,
	r *http.Request) {

	log.Printf("===============================================")
	log.Printf(" Method : DELETE ")
	log.Printf(" URL PATH : " + r.URL.Path[1:])

	vars := mux.Vars(r)

	log.Printf(" AFID  : %s", vars["afId"])
	log.Printf(" SUBSCRIPTION ID  : %s", vars["subscriptionId"])

	af, err := nef.nefGetAf(vars["afId"])

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err = af.afDeleteSubscription(vars["subscriptionId"])
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)

	logNef()
}

func logNef() {

	log.Printf("AF Count %+v", len(nef.afs))
	if len(nef.afs) > 0 {
		for key, value := range nef.afs {
			log.Printf(" AFKey : %+v, valAF_Id : %+v", key, value.afid)

			log.Printf("SUB Count for AF: %+v is %+v", value.afid, len(value.subs))
			for ks, vs := range value.subs {
				log.Printf("   SubKey : %+v, valSub_Id : %+v, ServiceId: %+v", ks, vs.subid, vs.ti.AfServiceID)
			}
		}
	}

}
