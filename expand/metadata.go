package expand

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/JFJun/stafi-substrate-go/utils"
	"github.com/huandu/xstrings"
	"github.com/stafiprotocol/go-substrate-rpc-client/types"
)

/*
对metadata进行扩展，添加一些实用的功能
由于大多数的波卡链都升级到了v11和v12，所以只对大于v11的链处理
*/
type MetadataExpand struct {
	meta *types.Metadata
	MV 	iMetaVersion
}
type iMetaVersion interface {
	GetCallIndex(moduleName, fn string) (callIdx string, err error)
	FindNameByCallIndex(callIdx string)(moduleName, fn string,err error)
}
func NewMetadataExpand(meta *types.Metadata)(*MetadataExpand,error){
	me :=new(MetadataExpand)
	me.meta = meta
	if meta.IsMetadataV11 {
		me.MV = newV11(meta.AsMetadataV11.Modules)
	}else if meta.IsMetadataV12 {
		me.MV = newV12(meta.AsMetadataV12.Modules)
	}else{
		return nil,errors.New("metadata version is not v11 or v12")
	}
	return me,nil
}

type v11 struct {
	module []types.ModuleMetadataV10
}

func (v v11) GetCallIndex(moduleName, fn string) (callIdx string, err error) {
	//避免指针为空
	defer func() {
		if errs := recover(); errs != nil {
			callIdx = ""
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	mi:=uint8(0)
	for _,mod :=range  v.module{
		if !mod.HasCalls {
			continue
		}
		if string(mod.Name) != moduleName {
			mi++
			continue
		}
		for ci,f:=range mod.Calls{
			if string(f.Name)==fn {
				return xstrings.RightJustify(utils.IntToHex(mi), 2, "0") + xstrings.RightJustify(utils.IntToHex(ci), 2, "0"),nil
			}
		}
	}
	return "", fmt.Errorf("do not find this call index")
}

func (v v11) FindNameByCallIndex(callIdx string) (moduleName, fn string, err error) {
	if len(callIdx)!=4 {
		return "", "", fmt.Errorf("call index length is not equal 4: length: %d",len(callIdx))
	}
	data,err:=hex.DecodeString(callIdx)
	if err != nil {
		return "","",fmt.Errorf("call index is not hex string")
	}
	mi:=int(data[0])
	ci:=int(data[1])
	for i,mod:=range v.module{
		if !mod.HasCalls {
			continue
		}
		if i==int(mi) {

			for j,call:=range mod.Calls{
				if j==int(ci){
					moduleName=string(mod.Name)
					fn = string(call.Name)
					return
				}
			}
		}
	}
	return "", "", fmt.Errorf("do not find this callInx info: %s",callIdx)
}

func newV11(module []types.ModuleMetadataV10)*v11{
	v:=new(v11)
	v.module = module
	return v
}
type v12 struct {
	module []types.ModuleMetadataV12
}

func (v v12) FindNameByCallIndex(callIdx string) (moduleName, fn string, err error) {
	if len(callIdx)!=4 {
		return "", "", fmt.Errorf("call index length is not equal 4: length: %d",len(callIdx))
	}
	data,err:=hex.DecodeString(callIdx)
	if err != nil {
		return "","",fmt.Errorf("call index is not hex string")
	}
	for _,mod:=range v.module{
		if !mod.HasCalls {
			continue
		}
		if mod.Index==data[0] {

			for j,call:=range mod.Calls{
				if j==int(data[1]){
					moduleName=string(mod.Name)
					fn = string(call.Name)
					return
				}
			}
		}
	}
	return "", "", fmt.Errorf("do not find this callInx info: %s",callIdx)
}

func (v v12) GetCallIndex(moduleName, fn string) (callIdx string, err error) {
	//避免指针为空
	defer func() {
		if errs := recover(); errs != nil {
			callIdx = ""
			err = fmt.Errorf("catch panic ,err=%v", errs)
		}
	}()
	for _,mod :=range  v.module{
		if !mod.HasCalls {
			continue
		}
		if string(mod.Name) != moduleName {

			continue
		}
		for ci,f:=range mod.Calls{
			if string(f.Name)==fn {
				return xstrings.RightJustify(utils.IntToHex(mod.Index),2,"0")+xstrings.RightJustify(utils.IntToHex(ci),2,"0"),nil
			}
		}
	}
	return "", fmt.Errorf("do not find this call index")
}

func newV12(module []types.ModuleMetadataV12)*v12{
	v:=new(v12)
	v.module = module
	return v
}


