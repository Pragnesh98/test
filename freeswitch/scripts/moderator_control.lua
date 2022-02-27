--[[
  The Initial Developer of the Original Code is
  
  Portions created by the Initial Developer are Copyright (C)
  the Initial Developer. All Rights Reserved.

  Contributor(s):

  moderator_control.lua --> conference controls 
--]]

	dofile("/usr/share/freeswitch/scripts/logger.lua");
	api = freeswitch.API();
	package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
	package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
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
	
	sounds_dir = session:getVariable("prompt_dir");
	freeswitch.consoleLog("notice", "[ConfAPI] :  Conference Controls DTMF ");
	digits = session:read(5, 10, "", 1000, "#");                                                                                                           
	session:consoleLog("info", "Got dtmf: ".. digits .."\n");
	
	if (tonumber(digits) == 6) then
		data = "non_moderator"
		freeswitch.consoleLog("notice", "[ConfAPI] :  Self Mute/Unmute");
		
		conference_name = session:getVariable("conference_name");
		CallUUID = session:getVariable("uuid");
		MemberID,is_moderator = getParticipantDetails(CallUUID, conference_name);
		freeswitch.consoleLog("notice", "[ConfAPI] :  MemberID : "..tostring(MemberID));
		
		mute_state = session:getVariable("mute_state");
		if (mute_state == "true" ) then
			session:setVariable("mute_state","false")
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " unmute " .. tostring(MemberID);
			response = api:executeString(cmd);
			
			freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
			freeswitch.consoleLog("notice", "[ConfAPI] :  You are UNMUTED Now");
			
		else
			session:setVariable("mute_state","true")
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " mute " .. tostring(MemberID);
			response = api:executeString(cmd);
			
			freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
			freeswitch.consoleLog("notice", "[ConfAPI] :  You are MUTED Now");
		end
		
	elseif (tonumber(digits) == 5) then
		data = "non_moderator"
		freeswitch.consoleLog("notice", "[ConfAPI] :  Group Mute/Unmute");
		
		conference_name = session:getVariable("conference_name");
		CallUUID = session:getVariable("uuid");

		group_mute_state = session:getVariable("group_mute_state");
		if (group_mute_state == "true" ) then
			session:setVariable("group_mute_state","false")
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " unmute all" ;
			response = api:executeString(cmd);
			
			freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
			freeswitch.consoleLog("notice", "[ConfAPI] :  Group unmuted");
			
		else
			session:setVariable("group_mute_state","true")
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " mute all";
			response = api:executeString(cmd);
			
			freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
			freeswitch.consoleLog("notice", "[ConfAPI] :  Group muted");
		end
		
	elseif (digits == "7") then
		conference_name = session:getVariable("conference_name");
		CallUUID = session:getVariable("uuid");
		freeswitch.consoleLog("notice", "[ConfAPI] :  Conference Lock/Unlock");
		MemberID,is_moderator, isLock = getParticipantDetails(CallUUID, conference_name);
		freeswitch.consoleLog("notice", "[ConfAPI] :  MemberID : "..tostring(MemberID));
		freeswitch.consoleLog("notice", "[ConfAPI] :  is_moderator : "..tostring(is_moderator));
	
		if (isLock == true or isLock == "true" ) then
			freeswitch.consoleLog("notice", "[ConfAPI] :  Conference Unlock");
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " unlock" ;
			response = api:executeString(cmd);
		else
			freeswitch.consoleLog("notice", "[ConfAPI] :  Conference Lock");
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " lock" ;
			response = api:executeString(cmd);
		end
		
	elseif (digits == "0*1") then
		freeswitch.consoleLog("notice", "[ConfAPI] :  Host Dial Out");
		
	elseif (digits == "0*3") then
		conference_name = session:getVariable("conference_name");
		CallUUID = session:getVariable("uuid");
		MemberID,is_moderator, isLock = getParticipantDetails(CallUUID, conference_name);
		freeswitch.consoleLog("notice", "[ConfAPI] :  MemberID : "..tostring(MemberID));
	
		if (isLock == true or isLock == "true" ) then
			freeswitch.consoleLog("notice", "[ConfAPI] :  Conference Unlock");
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " unlock" ;
			response = api:executeString(cmd);
		else
			freeswitch.consoleLog("notice", "[ConfAPI] :  Conference Lock");
			conference_name = session:getVariable("conference_name");
			cmd = "conference " .. conference_name .. " lock" ;
			response = api:executeString(cmd);
		end
	elseif (tonumber(digits) == 61) then
		data = "all"
		
		freeswitch.consoleLog("notice", "[ConfAPI] :  MUTE All Party");
		conference_name = session:getVariable("conference_name");
		cmd = "conference " .. conference_name .. " mute " .. data;
		response = api:executeString(cmd);
		
		freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));
			
	elseif (tonumber(digits) == 62) then
		data = "all"
		freeswitch.consoleLog("notice", "[ConfAPI] :  UNMUTE All Party");
		conference_name = session:getVariable("conference_name");
		cmd = "conference " .. conference_name .. " unmute " .. data;
		response = api:executeString(cmd);
		
		freeswitch.consoleLog("notice", "[ConfAPI] :  API cmd : "..tostring(cmd));

	elseif (tonumber(digits) == 0) then
		conference_name = session:getVariable("conference_name");
		CallUUID = session:getVariable("uuid");
		MemberID, speak_a, hear_a = getParticipantMuteStatus(CallUUID, conference_name);
		
		operator_request = session:getVariable("operator_request");
		if (operator_request == "true" ) then
			session:streamFile(sounds_dir.."/66.mp3");
			session:setVariable("operator_request","false")
			freeswitch.consoleLog("notice", "[ConfAPI] :  Your Operator Request is canceled");
			
			if tostring(hear_a):lower() == 'false' and  tostring(hear_a):lower() ~= nil then
				cmd = "conference "..tostring(conference_name) .." undeaf ".. MemberID
				reply = api:executeString(cmd);
				Logger.info ("[ConfAPI] : undeaf : "..reply.."");
				
				cmd = "conference "..tostring(conference_name) .." unmute ".. MemberID
				reply = api:executeString(cmd);
				Logger.info ("[ConfAPI] : Unmuted : "..reply.."");
			end
		else
			if tostring(hear_a):lower() == 'true' and tostring(hear_a):lower() ~= nil then 
				cmd = "conference "..tostring(conference_name) .." deaf ".. MemberID
				reply = api:executeString(cmd);
				Logger.info ("[ConfAPI] : Deafed : "..reply.."");

				cmd = "conference "..tostring(conference_name) .." mute ".. MemberID
				reply = api:executeString(cmd);
				Logger.info ("[ConfAPI] : Unmuted : "..reply.."");
			end
			session:streamFile(sounds_dir.."/65.mp3");
			session:setVariable("operator_request","true")

			cmd = "conference "..tostring(conference_name) .." dial sofia/internal/9974743321@18.216.103.124:9000"
			reply = api:executeString(cmd);
			freeswitch.consoleLog("notice", "[ConfAPI] :  API response ::  "..tostring(reply));
--  			split_string = Split(reply, ": [")
--  			freeswitch.consoleLog("notice", "[ConfAPI] :  split_string[2] :: "..tostring(split_string[2]));
-- 			DialStatus = Split(split_string[2], "]")
-- 			freeswitch.consoleLog("notice", "[ConfAPI] : DialStatus :: "..DialStatus);
-- 			local obCause = session:hangupCause()
-- 			freeswitch.consoleLog("info", "obSession:hangupCause() = " .. obCause )
				
-- 			session:execute("conference", "bridge:$1-${domain_name}@default:user/1000@${domain_name}");

			freeswitch.consoleLog("notice", "[ConfAPI] :  Your Operator Request.");
		end
		
	else 
		freeswitch.consoleLog("notice", "[ConfAPI] :  NO MATCH");
	end
	
