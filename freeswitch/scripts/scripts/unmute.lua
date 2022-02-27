
--get the argv values
	script_name = argv[0];

--options all, last, non_moderator, member_id
-- 	data = argv[1];

--prepare the api object
	api = freeswitch.API();
	package.cpath ="/usr/lib/x86_64-linux-gnu/lua/5.2/?.so;" .. package.cpath
	package.path = "/usr/share/lua/5.2/?.lua;" .. package.path
	xml = require "xml";
	
	debug["debug"] = false;
	
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
	
--get the session variables
	conference_name = session:getVariable("conference_name");
	freeswitch.consoleLog("notice", "[ConfAPI] : conference_name : ".. tostring(conference_name) .. "")
	CallUUID = session:getVariable("uuid");

	mute_state = session:getVariable("mute_state");
	if (mute_state == "true" ) then
		session:setVariable("mute_state","false")
		MemberID,is_moderator, islocked = getParticipantDetails(CallUUID, conference_name);
		freeswitch.consoleLog("notice", "[conference control] :  MemberID : "..tostring(MemberID));
			
	--send the conferenc mute command
		cmd = "conference " .. conference_name .. " unmute " ..tostring(MemberID);
		freeswitch.consoleLog("notice", "[ConfAPI] : cmd : ".. tostring(cmd) .. "")
		response = api:executeString(cmd);
		freeswitch.consoleLog("notice", "[ConfAPI] : response : ".. tostring(response) .. "")
	end
