        script_dir = freeswitch.getGlobalVariable("script_dir")
        dofile(script_dir.."/logger.lua");
        if(script_dir == "/usr/local/freeswitch/scripts") then
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/local/share/lua/5.1/?.lua;" .. package.path
        else
                package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
                package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        end

        api = freeswitch.API();
--      package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
--      package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
        xml = require "xml";

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

        freeswitch.consoleLog("notice", "[ConfAPI] :  Self Mute/Unmute");

        conference_name = session:getVariable("conference_name");
        CallUUID = session:getVariable("uuid");
        MemberID, speak_a, hear_a = getParticipantMuteStatus(CallUUID, conference_name);
        freeswitch.consoleLog("notice", "[ConfAPI] :  MemberID : "..tostring(MemberID));

	cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
	participant_mute = session:getVariable("participant_mute");
	--participant_mute = "false";
	freeswitch.consoleLog("notice", "[ConfAPI] :  particiapnt_mute : "..tostring(participant_mute));

--if(tostring(participant_mute) == "f") then
        mute_status = argv[3]
        if (tostring(speak_a) == "false") then
        --if tostring(speak_a):lower() == 'false' and  tostring(speak_a):lower() ~= nil then
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " unmute " .. tostring(MemberID);
                response = api:executeString(cmd);

                freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
                freeswitch.consoleLog("notice", "[ConfAPI] :  You are UNMUTED Now");
        else
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " mute " .. tostring(MemberID);
                response = api:executeString(cmd);

                freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
                freeswitch.consoleLog("notice", "[ConfAPI] :  You are MUTED Now");
        end
--end
