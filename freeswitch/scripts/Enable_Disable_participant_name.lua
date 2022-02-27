	script_dir = freeswitch.getGlobalVariable("script_dir")
	dofile(script_dir.."/logger.lua");
	if(script_dir == "/usr/local/freeswitch/scripts") then
		package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
		package.path = "/usr/local/share/lua/5.1/?.lua;" .. package.path
	else
		package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
		package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
	end
        require "resources.functions.base64";
        require "resources.functions.trim";
        require "resources.functions.file_exists";
        require "resources.functions.explode";
        require "resources.functions.format_seconds";
        require "resources.functions.mkdir";
        local json = require "resources.functions.lunajson"
	xml = require "xml";
	
	api = freeswitch.API();
	debug["debug"] = false;
	
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

	function all_participant_uuids(conf_name)
		result = api:executeString("conference "..conf_name .. " xml_list")
		result = result:gsub("<variables></variables>", "")
		xmltable = xml.parse(result)
		
		if (debug["debug"]) then
			Logger.debug ("[ConfAPI] : xml_list ".. tostring(xmltable) .. "")
		end
		
		alluuids =  {};

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
								table.insert(alluuids,subtag[j][k][1])
								Logger.debug ("[ConfAPI] : UUID : ".. tostring(subtag[j][k][1]) .. "")
							end
						end
					end 
				end
			end
			
			i = i+1;
		end
	
		return alluuids
	end
local function Enable_Disable(session, conference_name, CallUUid)	
	allUUID = {};

        UrL = freeswitch.getGlobalVariable("UrL")
	Logger.notice ("[ConfAPI] : Enable/Disable participant name announcement")
	sounds_dir = session:getVariable("prompt_dir");
--	session:execute("playback", sounds_dir.."/103.wav");
	conference_name = session:getVariable("conference_name");
	CallUUID = session:getVariable("uuid");
	MemberID, speak_a, hear_a = getParticipantMuteStatus(CallUUID, conference_name);
	
	conference_id = session:getVariable("conf_id")

	destination_number = session:getVariable("destination_number");
	pin_number = session:getVariable("pin_number")
        APIcmd = "curl --location --request GET '"..tostring(UrL).."did/verify-pin/"..tostring(destination_number).."/"..tostring(pin_number).."'"
        Logger.info ("[MindBridge] : Conf PIN verification APIcmd : "..tostring(APIcmd).."");
        curl_response = api:execute("system", APIcmd);
        Logger.debug ("[MindBridge] : Conf PIN verification API Response : "..tostring(curl_response).."");

        local encode = json.decode (curl_response)
        local MessageRes = encode['message']
        local Msg = encode['Msg']
        if(MessageRes ~= "Not Found" or Msg ~= "Failed") then
		enable_disable  = encode['did_conference_map']['enable_disable']
                Logger.debug ("[MindBridge] : enable_disable : " ..tostring(enable_disable));
		
	end
	Logger.debug ("[MindBridge] : Enable_Disable : " ..tostring(enable_disable).." conference_id: "..tostring(conference_id).."");

if tostring(enable_disable) == "t" then
	--allUUIDs = all_participant_uuids(conference_name)
        APIcmd = "curl --location --request GET '"..tostring(UrL).."enable/"..tostring(conference_id).."/f'"
        Logger.info ("[MindBridge] : APIcmd : "..tostring(APIcmd).."");
        curl_response = api:execute("system", APIcmd);
        Logger.debug ("[MindBridge] : API Response : "..tostring(curl_response).."");
	session:execute("playback", sounds_dir.."/104.wav");
else
        APIcmd = "curl --location --request GET '"..tostring(UrL).."enable/"..tostring(conference_id).."/t'"
        Logger.info ("[MindBridge] : APIcmd : "..tostring(APIcmd).."");
        curl_response = api:execute("system", APIcmd);
        Logger.debug ("[MindBridge] : API Response : "..tostring(curl_response).."");
	session:execute("playback", sounds_dir.."/103.wav");
end

	--session:setVariable("Enable", "true");
		cnt = 0
		session:sleep(1000);
		Keypad = require 'keypad_commands-loop'
		Keypad(session, conference_name, aleg_uuid,cnt)
 end

 return Enable_Disable
	
