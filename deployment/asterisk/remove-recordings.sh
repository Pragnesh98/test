#!/bin/bash
find /var/spool/asterisk/recording/  -name "*.wav" -type f -mtime +2 -exec rm -f {} \;

#Delete TTS Files
find /var/lib/asterisk/sounds/en/ -name "*wavenet*.mp3" -type f -mtime +5 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*wavenet*.sln" -type f -mtime +5 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*polly*.mp3" -type f -mtime +2 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*polly*.sln" -type f -mtime +2 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*azure*.mp3" -type f -mtime +2 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*azure*.alaw" -type f -mtime +2 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*azure*.wav" -type f -mtime +2 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*azure*.sln" -type f -mtime +2 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*url*.sln" -type f -mtime +2 -exec rm -f {} \;

find /var/lib/asterisk/sounds/en/ -name "*url*.mp3" -type f -mtime +2 -exec rm -f {} \;
