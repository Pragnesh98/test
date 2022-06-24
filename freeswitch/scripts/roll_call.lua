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
--local function roll_call (session, conference_name, CallUUID)
	
	allUUID = {};

	Logger.notice ("[ConfAPI] : Announce Names")
	sounds_dir = session:getVariable("prompt_dir");
	conference_name = session:getVariable("conference_name");
	CallUUID = session:getVariable("uuid");
	MemberID, speak_a, hear_a = getParticipantMuteStatus(CallUUID, conference_name);
	
	allUUIDs = all_participant_uuids(conference_name)
	
  for _,calluuid in pairs(allUUIDs) do
    --if(CallUUID ~= calluuid) then
		Logger.notice ("[ConfAPI] : calluuid :: "..tostring(calluuid))
		cmd = "conference "..tostring(conference_name).." play /tmp/conference-"..calluuid..".mp3 "..MemberID
		response = api:executeString(cmd);
		session:sleep(500);
	end
-- end