--[[
  The Initial Developer of the Original Code is

  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  keypad_commands.lua --> conference controls
--]]



        script_dir = freeswitch.getGlobalVariable("script_dir")
        dofile(script_dir.."/logger.lua");
        if(script_dir == "/usr/local/freeswitch/scripts") then
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/local/share/lua/5.1/?.lua;" .. package.path
        else
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        end
        xml = require "xml";
        debug["debug"]=false

        api = freeswitch.API();

        function getParticipantDetails(unique_id,conf_name)
                id, hear_value, speak_value = nil, nil, nil;
                is_moderator = false;
                isLock = false;
                result = api:executeString("conference "..conf_name .. " xml_list")
                result = result:gsub("<variables></variables>", "")

                if(result == nil or result == "nil" or result == "") then
                        return 0,false, isLock;
                end

                xmltable = xml.parse(result)
                for i=1,#xmltable do
                        if xmltable[i].tag then
                                local tag = xmltable[i]
                                local subtag = tag[i]
                                isLock = tag.attr.locked
                                if subtag.tag == "members" then
                                        if (debug["debug"]) then
                                                freeswitch.consoleLog("notice", "[ConfAPI] : subtag : ".. tostring(subtag) .. "")
                                        end
                                        for j = 1,#subtag do
                                                for k = 1,#subtag[j] do
                                                        if (debug["debug"]) then
                                                                freeswitch.consoleLog("notice", "[ConfAPI] : subtag[j][k] : ".. tostring(subtag[j][k]) .. "")
                                                        end
                                                        if( subtag[j][k] ~= nil) then
                                                                if tostring(subtag[j][k].tag) == 'uuid' then
                                                                        if  subtag[j][k][1] == unique_id then
                                                                                if (debug["debug"]) then
                                                                                        freeswitch.consoleLog("notice", "[ConfAPI] : Log  Data "..unique_id.."\tConference_Name : "..conf_name.."\nFS RESPONSE :  "..result.."");
                                                                                end
                                                                                s1 = subtag[j]
                                                                                for i, j in ipairs(s1) do
                                                                                        if( string.find(tostring(j),"<id>") ~= nil )then
                                                                                                id = tostring(j):match("id>(.-)</id>")
                                                                                        elseif(string.find(tostring(j),"<is_moderator>") ~= nil ) then
                                                                                                is_moderator = tostring(j):match("is_moderator>(.-)</is_moderator>")
                                                                                        end
                                                                                end
                                                                        end
                                                                end
                                                        end
                                                end
                                        end
                                end
                        end
                        i = i+1;
                end

                freeswitch.consoleLog("notice", "[ConfAPI] : ID : "..tostring(id).." IS_MODERATOR : ["..tostring(is_moderator).."] ISLOCK ["..tostring(isLock).."]");
                return id, is_moderator, isLock
        end
        freeswitch.consoleLog("notice", "[ConfAPI] :  List available keypad commands");
--        sounds_dir = session:getVariable("prompt_dir");
       sounds_dir ="/usr/local/freeswitch/prompt"
        conference_name = session:getVariable("conference_name");
        CallUUID = session:getVariable("uuid");

--local function Keypad(session, conference_name, CallUUID)
        MemberID,is_moderator, islocked = getParticipantDetails(CallUUID, conference_name);
        freeswitch.consoleLog("notice", "[conference control] :  MemberID : "..tostring(MemberID));

        cmd = "conference "..tostring(conference_name).." moh off"
        response = api:executeString(cmd);
	user_number = session:playAndGetDigits(2, 4, 1,1000, "#", "/usr/local/freeswitch/prompt/leaderprompt.wav", "", "");

	Logger.info("[MindBridge] : user_input: " .. tostring(user_number) .. "");
	aleg_uuid = session:getVariable("uuid");
	
	if(user_number == "*0") then
		request_type = "individual";
		Operator = require 'operator_request'
		Operator(session, conference_name, aleg_uuid, request_type)
	elseif(user_number == "*1") then
		session:execute("playback", sounds_dir.."/leaderprompt.wav");
                 Rejoin = require 'keypad_commands'
                Rejoin(session, conference_name, aleg_uuid)
		--Dialout = require 'dialout'
		--Dialout(session, conference_name, aleg_uuid)
	elseif(user_number == "*2") then
		--cnt =0
		Participant_Name = require 'participants_name'
		Participant_Name(session, conference_name, aleg_uuid)
		--Current_Participant = require 'current_participant'
		--Current_Participant(session, conference_name, aleg_uuid, cnt)
	elseif(user_number == "*3") then
		lock_unlock = require 'lock_unlock03'
		lock_unlock(session, conference_name, aleg_uuid)
	elseif(user_number == "*4") then
		Conf_Call = require 'conf_call'
		Conf_Call(session, conference_call, aleg_uuid)
	elseif(user_number == "*5") then
		Enable_Disable = require 'Enable_Disable_participant_name'
		Enable_Disable(session, conference_call, aleg_uuid)
	elseif(user_number == "*6") then
		cnt = 0
		New = require 'new-1'
                New(session , conference_call, aleg_uuid, cnt, 0)
		--NotJoin();
	elseif(user_number == "*7") then
		session:execute("playback", sounds_dir.."/option_not_available.wav");
		NotJoin();
	elseif(user_number == "*8") then
		session:execute("playback", sounds_dir.."/option_not_available.wav");
		Notjoin();
        elseif(user_number == "*9") then
		Rejoin = require 'rejoin'
		Rejoin(session, conference_name, aleg_uuid)
	else
		cnt = 0
		session:sleep(1000);
		Keypad = require 'keypad_commands-loop'
		Keypad(session, conference_name, aleg_uuid, cnt)

	end

--end

--return Keypa
function NotJoin()
	cnt = 0
	session:sleep(1000);
	Keypad = require 'keypad_commands-loop'
	Keypad(session, conference_name, aleg_uuid, cnt)

end
	-- Keypad(session, conference_name, CallUUID)
