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

        api = freeswitch.API();
        debug["debug"] = false;

        function Split(s, delimiter)
                result = {};
                for match in (s..delimiter):gmatch("(.-)"..delimiter) do
                        table.insert(result, match);
                end
                return result;
        end

        function getParticipantMuteStatus(unique_id,conf_name)
                id, hear_value, speak_value = nil, nil, nil;

                result = api:executeString("conference "..conf_name .. " xml_list")
                result = result:gsub("<variables></variables>", "")
                xmltable = xml.parse(result)

                if (debug["debug"]) then
                        Logger.debug ("[ConfAPI] : xml_list ".. tostring(xmltable) .. "")
                end

                for i=1,#xmltable do
                        if xmltable[i].tag then
                                local tag = xmltable[i]
                                local subtag = tag[i]

                                if subtag.tag == "members" then
                                        if (debug["debug"]) then
                                                Logger.debug ("[ConfAPI] : subtag : ".. tostring(subtag) .. "")
                                        end

                                        for j = 1,#subtag do
                                                for k = 1,#subtag[j] do
                                                        if (debug["debug"]) then
                                                                Logger.debug ("[ConfAPI] : subtag[j][k] : ".. tostring(subtag[j][k].tag) .. "")
                                                        end

                                                        if subtag[j][k].tag == 'uuid' then
                                                                if  subtag[j][k][1] == unique_id then

                                                                        if (debug["debug"]) then
                                                                                Logger.debug ("[ConfAPI] : Log  Data "..unique_id.."\tConference_Name : "..conf_name.."\nFS RESPONSE :  "..result.."");
                                                                        end

                                                                        s1 = subtag[j]
                                                                        for i, j in ipairs(s1) do
                                                                                if( string.find(tostring(j),"<id>") ~= nil )then
                                                                                        id = tostring(j):match("id>(.-)</id>")
                                                                                elseif(string.find(tostring(j),"<can_hear>") ~= nil ) then

                                                                                        hear_value = tostring(j):match("can_hear>(.-)</can_hear>")
                                                                                        speak_value = tostring(j):match("can_speak>(.-)</can_speak>")

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

                Logger.info ("[ConfAPI] : ID : "..tostring(id).." SPEAK : ["..tostring(speak_value).."] HEAR : ["..tostring(speak_value).."]");

                return id, speak_value, hear_value
        end
        request_type = argv[1];

        Logger.notice ("[ConfAPI] : conference dialout request")
	sounds_dir =freeswitch.getGlobalVariable("prompt_dir")
	conference_name = argv[1]
	CallUUID = argv[2]
        MemberID, speak_a, hear_a = getParticipantMuteStatus(CallUUID, conference_name);

	cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));

        --session:setVariable("hangup_after_bridge","false");
        if(request_type == "conference") then
                operator_request = session:getVariable("operator_request");
                if (operator_request == "true" ) then
                        session:streamFile(sounds_dir.."/operatorreq.wav");
                        session:setVariable("operator_request","false")
                        freeswitch.consoleLog("notice", "[ConfAPI] :  Your Operator Request is canceled");

        --[[            if tostring(hear_a):lower() == 'false' and  tostring(hear_a):lower() ~= nil then
                                cmd = "conference "..tostring(conference_name) .." undeaf ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : undeaf : "..reply.."");

                                cmd = "conference "..tostring(conference_name) .." unmute ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : Unmuted : "..reply.."");
                        end ---]]
                else
--[[                    if tostring(hear_a):lower() == 'true' and tostring(hear_a):lower() ~= nil then
                                cmd = "conference "..tostring(conference_name) .." deaf ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : Deafed : "..reply.."");

                                cmd = "conference "..tostring(conference_name) .." mute ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : Unmuted : "..reply.."");
                        end
--]]
                        session:streamFile(sounds_dir.."/operatorreq.wav");
                        session:setVariable("operator_request","true")

                        sofia_str = "{origination_caller_id_number='+912230912202',origination_caller_id_name='+912230912202'}sofia/internal/+918849145548@80.77.12.162:5060;fs_path=sip:80.77.12.162,{origination_caller_id_number='+912230912202',origination_caller_id_name='+912230912202'}sofia/internal/+918849145548@80.77.12.162:5060;fs_path=sip:80.77.12.162"
                        cmd = "conference "..tostring(conference_name).." dial "..sofia_str.." +912230912202 +912230912202"
                        response = api:executeString(cmd);

                        freeswitch.consoleLog("notice", "[ConfAPI] :  Your Operator Request.");
                end
        else
                operator_request = session:getVariable("operator_request_in");
                if (operator_request == "true" ) then
                        session:streamFile(sounds_dir.."/66.mp3");
                        session:setVariable("operator_request_in","false")
                        freeswitch.consoleLog("notice", "[ConfAPI] :  Your Operator Request is canceled");
                        dialout_uuid = session:getVariable("dialout_uuid");
                        cmd = "uuid kill "..tostring(dialout_uuid)
                        originate_uuid = api:executeString(cmd);

--[[                    if tostring(hear_a):lower() == 'false' and  tostring(hear_a):lower() ~= nil then
                                cmd = "conference "..tostring(conference_name) .." undeaf ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : undeaf : "..reply.."");

                                cmd = "conference "..tostring(conference_name) .." unmute ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : Unmuted : "..reply.."");
                        end
--]]
                else
--[[                    if tostring(hear_a):lower() == 'true' and tostring(hear_a):lower() ~= nil then
                                cmd = "conference "..tostring(conference_name) .." deaf ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : Deafed : "..reply.."");

                                cmd = "conference "..tostring(conference_name) .." mute ".. MemberID
                                reply = api:executeString(cmd);
                                Logger.info ("[ConfAPI] : Unmuted : "..reply.."");
                        end
--]]

                        cmd = "create_uuid"
                        originate_uuid = api:executeString(cmd);
                        Logger.info ("[ConfAPI] : originate UUID : "..originate_uuid.."");
                        session:streamFile(sounds_dir.."/operatorreq.wav");
                        session:setVariable("operator_request_in","true")
                        session:setVariable("dialout_uuid",originate_uuid)

			session:execute("playback", sounds_dir.."/option_not_available.wav");
                        --session:execute("bridge","{originate_uuid="..originate_uuid..",origination_caller_id_number='+912230912202',origination_caller_id_name='+912230912202'}sofia/internal/+918849145548@80.77.12.162:5060;fs_path=sip:80.77.12.162");
                        --session:setVariable("operator_request_in","false")
                        conf_id = session:getVariable("conf-id");
                        if(conf_id) then
                                session:execute("conference", conf_id);
                        end
                        freeswitch.consoleLog("notice", "[ConfAPI] :  Your Operator Request.");
                end
end
--
--	--cnt = 0
--		session:sleep(1000);
--		Keypad = require 'keypad_commands-loop'
--		Keypad(session, conference_name, aleg_uuid, cnt)


        --api:executeString("msleep 1000")
	--reply = api:executeString("luarun keypad_commands.lua "..tostring(conference_name).." "..tostring(aleg_uuid))
