// Cynhyrchwyd y ffeil hon yn awtomatig. PEIDIWCH Â MODIWL
// This file is automatically generated. DO NOT EDIT
import {config} from '../models';
import {state_manager} from '../models';

export function GetAppConfig():Promise<config.AppConfig>;

export function GetCurrentSnapshot():Promise<state_manager.Snapshot>;

export function StartCapture(arg1:string,arg2:number,arg3:string,arg4:string):Promise<void>;

export function StopCapture():Promise<void>;
