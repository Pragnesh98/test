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
                                                                                        elseif(string.find(tostring(j),"<is_guest>") ~= nil ) then
                                                                                                is_moderator = tostring(j):match("is_guest>(.-)</is_guest>")
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

                freeswitch.consoleLog("notice", "[ConfAPI] : ID : "..tostring(id).." IS_guest : ["..tostring(is_moderator).."] ISLOCK ["..tostring(isLock).."]");
                return id, is_guest, isLock
        end

        freeswitch.consoleLog("notice", "[ConfAPI] :  Conference lock/Unlock");
        sounds_dir = session:getVariable("prompt_dir");

        conference_name = session:getVariable("conference_name");
        CallUUID = session:getVariable("uuid");

	MemberID,is_moderator, isLock = getParticipantDetails(CallUUID, conference_name);
        freeswitch.consoleLog("notice", "[ConfAPI] :  MemberID : "..tostring(MemberID));
        freeswitch.consoleLog("notice", "[ConfAPI] :  is_moderator : "..tostring(is_moderator));


	cmd = "conference " .. conference_name .. " moh off"
	response = api:executeString(cmd);
	freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
	freeswitch.consoleLog("notice", "[ConfAPI] :  isLock : "..tostring(isLock));

        mute_status = argv[3]
        --if (tostring(speak_a) == "false") then
        --if tostring(speak_a):lower() == 'false' and  tostring(speak_a):lower() ~= nil then
	if (isLock == true or isLock == "true" ) then
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " unlock " .. tostring(MemberID);
                response = api:executeString(cmd);
                 

                freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
                freeswitch.consoleLog("notice", "[ConfAPI] :  The conference is now unlocked");
        else
                conference_name = session:getVariable("conference_name");
                cmd = "conference " .. conference_name .. " lock " .. tostring(MemberID);
                response = api:executeString(cmd);
                   

                freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
                freeswitch.consoleLog("notice", "[ConfAPI] :  The conference is now locked");
        end
            

         

--	conference_name = session:getVariable("conference_name");
--	cmd = "conference "..tostring(conference_name).." play "..sounds_dir.."/leaderprompt.wav "..MemberID
--	response = api:executeString(cmd);
  --            freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));


