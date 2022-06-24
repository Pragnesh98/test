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
	
	function play_member_count(announce_count, member_count, member_count_prompt)
		--play member count
		if (announce_count == "true") then
			if (member_count == "2") then
			--there is one other member in this conference
				session:execute("playback", "conference/conf-one_other_member_conference.wav");
			elseif (member_count == "1") then
-- 				if (wait_mod == "true" and member_type ~= "moderator") then
 					session:execute("playback", member_count_prompt);
-- 				end
				--conference profile defines the alone sound file
			else
				--say the count
				session:execute("say", "en number pronounced "..member_count);
				--members in this conference
				session:execute("playback", "conference/conf-members_in_conference.wav");
			end
		end
	end
	
	conference_name = session:getVariable("conference_name");
	CallUUID = session:getVariable("uuid");
	sounds_dir = session:getVariable("prompt_dir");

	cmd = "conference "..tostring(conference_name).." list count";
	member_count = api:executeString(cmd);
	if string.match(member_count, "not found") then
		member_count = "0";
	end
	freeswitch.consoleLog("notice", "[ConfAPI] :  member_count : "..tostring(member_count));
	play_member_count("true", member_count, sounds_dir.."/45.mp3")
	