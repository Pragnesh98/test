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

--function hangup_function_name()
--	exit()
--end
Cnt = 0;
function NotJoin(user_number, Cnt, counter)

	session:sleep(1000);
        Logger.info("[MindBridge] : user_number : " .. tostring(user_number) .. "");
	if tostring(user_number) == "*3" then
		Cnt = Cnt+1
	else
		Cnt = 0;
	end
	New = require 'new-1'
	New(session, conference_name, aleg_uuid, Cnt, counter)
        	Logger.info("[MindBridge] : Cnt : " .. tostring(Cnt) .. "");

end

local function New(session, conference_name, CallUUID, Cnt, counter)
        	Logger.info("[MindBridge] : Cnt : " .. tostring(Cnt) .. ", counter: "..tostring(counter));
	
	--hangup = session:getVariable("hangup")
	--result = session.setHangupHook(hangup_function_name);
        freeswitch.consoleLog("notice", "[ConfAPI] :  List available keypad commands new-1");
--        sounds_dir = session:getVariable("prompt_dir");
       sounds_dir ="/usr/local/freeswitch/prompt"
        conference_name = session:getVariable("conference_name");
        --CallUUID = session:getVariable("uuid");
        
	MemberID,is_moderator, islocked = getParticipantDetails(CallUUID, conference_name);
        freeswitch.consoleLog("notice", "[conference control] :  MemberID : "..tostring(MemberID));

        cmd = "conference "..tostring(conference_name).." moh off"
        response = api:executeString(cmd);
	Logger.notice ("[ConfAPI] : Participant manage_menu")
       -- sounds_dir = session:getVariable("prompt_dir");
        user_number = session:playAndGetDigits(2, 4, 1,1000, "#", "/usr/local/freeswitch/prompt/Diff_ivr.wav", "", "");

        Logger.info("[MindBridge] : user_input: " .. tostring(user_number) .. "");
        aleg_uuid = session:getVariable("uuid");

        if(user_number == "*1") then
                cmd = "conference "..tostring(conference_name).." list count";
                member_count = api:executeString(cmd);
                if string.match(member_count, "not found") then
                        member_count = "0";
                end
                if (tonumber(member_count) == 2) then
                       cnt = 0
                elseif (tonumber(member_count) >= 3) then
                       cnt = 1
                end

                 Previous_Participant = require 'previous_participant'
                 Previous_Participant(session, conference_name, CallUUID, cnt)
               	NotJoin(user_number, Cnt, counter);
		 --reply = api:executeString("luarun roll_call.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
         elseif(user_number == "*2") then
                 cnt = 0
                 Current_Participant = require 'current_participant'
                 Current_Participant(session, conference_name, CallUUID, cnt)
                 NotJoin(user_number, Cnt, counter);
		 --Participant_Name = require 'participants_name1'
                 --Participant_Name(session, conference_name, aleg_uuid)
                --reply = api:executeString("luarun roll_call1.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
        elseif(user_number == "*3") then
        	Logger.info("[MindBridge] : Cnt : " .. tostring(Cnt) .. "");

                Next_Participant = require 'next_participant'
                 Next_Participant(session, conference_name, aleg_uuid, Cnt)
            	NotJoin(user_number, Cnt, counter);
		 --    reply = api:executeString("luarun next participants.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
        elseif(user_number == "*4") then
		session:execute("playback", sounds_dir.."/option_not_available.wav");
            	    NotJoin(user_number, Cnt, counter);
                --reply = api:executeString("luarun disconnect_private_call.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
         elseif(user_number == "*5") then
         	    session:execute("playback", sounds_dir.."/confirm_or_cancel_disconnection.wav");
            	    --[[NotJoin(user_number, Cnt);
		    if(user_number == "*1") then
			    session:execute("playback", sounds_dir.."/option_not_available.wav");
		    elseif(user_number == "*2") then
			    session:execute("playback", sounds_dir.."/action_cancled.wav");
		    NotJoin(user_number, Cnt);
	            else
	                end

                      --]]

               -- reply = api:executeString("luarun disconnect_private_call.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
        elseif(user_number == "*6") then
                Counts = require 'participants_counts'
                Counts(session, conference_name, aleg_uuid, Cnt)
            	NotJoin(user_number, Cnt, counter);
                   --session:execute("playback", sounds_dir.."/option_not_available.wav");
        elseif(user_number == "*7") then
		     Roll_Call = require 'participants_name'
                     Roll_Call(session, conference_name, aleg_uuid)
            	    NotJoin(user_number, Cnt, counter);
                    --session:execute("playback", sounds_dir.."/option_not_available.wav");
        elseif(user_number == "*8") then
                    session:execute("playback", sounds_dir.."/option_not_available.wav");
            	    NotJoin(user_number, Cnt, counter);
        elseif(user_number == "*9") then
                --    session:execute("playback", sounds_dir.."/rejoin.wav");
                 	 Rejoin = require 'rejoin'
                Rejoin(session, conference_name, aleg_uuid)
	else
		if counter == 20 then
			exit(10);
		end
		counter = counter + 1;
		NotJoin(user_number, Cnt, counter);
		    --  reply = api:executeString("luarun rejoin.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
           end
               --cnt = 0
		--t = cnt+1
                --ssion:sleep(1000);
                --w = require 'new-1'
                --New(session, conference_name, aleg_uuid,cnt)

end

return New

