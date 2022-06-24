#include <switch.h>

#define AMD_PARAMS (2)
#define AMD_SYNTAX "<uuid> <command>"

SWITCH_MODULE_SHUTDOWN_FUNCTION(mod_amd_shutdown);
SWITCH_MODULE_LOAD_FUNCTION(mod_amd_load);
SWITCH_MODULE_DEFINITION(mod_amd, mod_amd_load, mod_amd_shutdown, NULL);
SWITCH_STANDARD_APP(amd_start_function);

static struct {
	uint32_t initial_silence;
	uint32_t greeting;
	uint32_t after_greeting_silence;
	uint32_t total_analysis_time;
	uint32_t minimum_word_length;
	uint32_t between_words_silence;
	uint32_t maximum_number_of_words;
	uint32_t silence_threshold;
	uint32_t maximum_word_length;
} globals;

/* don't forget to update avmd_events_str table if you modify this */
enum amd_event
{
    AMD_EVENT_MACHINE = 0,
    AMD_EVENT_SESSION_START = 1,
    AMD_EVENT_SESSION_STOP = 2
};
/* This array MUST be NULL terminated! */
const char* amd_events_str[] = {
    [AMD_EVENT_MACHINE] =             "amd::machine",
    [AMD_EVENT_SESSION_START] =    "amd::start",
    [AMD_EVENT_SESSION_STOP] =     "amd::stop",
    NULL                                            /* MUST be last and always here */
};

#define AMD_CHAR_BUF_LEN 20u
#define AMD_BUF_LINEAR_LEN 160u

typedef enum {
    MACHINE_DETECTED,
    MACHINE_NOTDETECTED
} amd_beep_state_t;


static switch_status_t amd_register_all_events(void);

static void amd_unregister_all_events(void);

static void amd_reloadxml_event_handler(switch_event_t *event);

static void amd_fire_event(enum amd_event type, switch_core_session_t *fs_s, double freq, double v_freq, double amp, double v_amp, amd_beep_state_t machine_status, uint8_t info,
        switch_time_t detection_start_time, switch_time_t detection_stop_time, switch_time_t start_time, switch_time_t stop_time, uint8_t resolution, uint8_t offset, uint8_t idx);

static switch_xml_config_item_t instructions[] = {
	SWITCH_CONFIG_ITEM(
		"initial_silence",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.initial_silence,
		(void *) 2500,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"greeting",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.greeting,
		(void *) 1500,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"after_greeting_silence",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.after_greeting_silence,
		(void *) 800,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"total_analysis_time",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.total_analysis_time,
		(void *) 5000,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"min_word_length",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.minimum_word_length,
		(void *) 100,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"between_words_silence",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.between_words_silence,
		(void *) 50,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"maximum_number_of_words",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.maximum_number_of_words,
		(void *) 3,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"maximum_word_length",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.maximum_word_length,
		(void *)5000,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM(
		"silence_threshold",
		SWITCH_CONFIG_INT,
		CONFIG_RELOADABLE,
		&globals.silence_threshold,
		(void *) 256,
		NULL, NULL, NULL),

	SWITCH_CONFIG_ITEM_END()
};

static switch_status_t do_config(switch_bool_t reload)
{
	memset(&globals, 0, sizeof(globals));

	if (switch_xml_config_parse_module_settings("amd.conf", reload, instructions) != SWITCH_STATUS_SUCCESS) {
		return SWITCH_STATUS_FALSE;
	}

	return SWITCH_STATUS_SUCCESS;
}

static switch_status_t amd_register_all_events(void) {
    size_t idx = 0;
    const char *e = amd_events_str[0];
    while (e != NULL)
    {
        if (switch_event_reserve_subclass(e) != SWITCH_STATUS_SUCCESS) {
            switch_log_printf(SWITCH_CHANNEL_LOG, SWITCH_LOG_ERROR, "Couldn't register subclass [%s]!\n", e);
            return SWITCH_STATUS_TERM;
        }
        ++idx;
        e = amd_events_str[idx];
    }
    return SWITCH_STATUS_SUCCESS;
}

static void amd_unregister_all_events(void) {
    size_t idx = 0;
    const char *e = amd_events_str[0];
    while (e != NULL)
    {
        switch_event_free_subclass(e);
        ++idx;
        e = amd_events_str[idx];
    }
    return;
}

static void amd_fire_event(enum amd_event type, switch_core_session_t *fs_s, double freq, double v_freq, double amp, double v_amp,amd_beep_state_t machine_status, uint8_t info,
        switch_time_t detection_start_time, switch_time_t detection_stop_time, switch_time_t start_time, switch_time_t stop_time, uint8_t resolution, uint8_t offset, uint8_t idx) {
    int res;
    switch_event_t      *event;
    switch_time_t       total_time;
    switch_status_t     status;
    switch_event_t      *event_copy;
    char                buf[AMD_CHAR_BUF_LEN];

    status = switch_event_create_subclass(&event, SWITCH_EVENT_CUSTOM, amd_events_str[type]);
    if (status != SWITCH_STATUS_SUCCESS) {
        return;
    }
    switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Unique-ID", switch_core_session_get_uuid(fs_s));
    switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Call-command", "amd");
    switch (type)
    {
        case AMD_EVENT_MACHINE:
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Machine-Status", "DETECTED");
            /*res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%f", freq);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Frequency truncated [%s], [%d] attempted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Frequency", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Frequency", buf);

            res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%f", v_freq);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Error, truncated [%s], [%d] attempeted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Frequency-variance", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Frequency-variance", buf);
                                                                                                 
	    res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%f", amp);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Amplitude truncated [%s], [%d] attempted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Amplitude", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Amplitude", buf);

            res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%f", v_amp);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Error, truncated [%s], [%d] attempeted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Amplitude-variance", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Amplitude-variance", buf);

            detection_time = detection_stop_time - detection_start_time;
            res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%" PRId64 "", detection_time);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Detection time truncated [%s], [%d] attempted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detection-time", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detection-time", buf);

            res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%u", resolution);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Error, truncated [%s], [%d] attempeted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detector-resolution", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detector-resolution", buf);

            res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%u", offset);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Error, truncated [%s], [%d] attempeted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detector-offset", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detector-offset", buf);

            res = snprintf(buf, AVMD_CHAR_BUF_LEN, "%u", idx);
            if (res < 0 || res > AVMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Error, truncated [%s], [%d] attempeted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detector-index", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Detector-index", buf);*/
            break;

        case AMD_EVENT_SESSION_START:
            res = snprintf(buf, AMD_CHAR_BUF_LEN, "%" PRId64 "", start_time);
            if (res < 0 || res > AMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Start time truncated [%s], [%d] attempted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Start-time", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Start-time", buf);
            break;

        case AMD_EVENT_SESSION_STOP:
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Machine-Status", machine_status == MACHINE_DETECTED ? "DETECTED" : "NOTDETECTED");
            if (info == 0) {
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Stop-status", "ERROR (AMD SESSION OBJECT NOT FOUND IN MEDIA BUG)");
            }
            total_time = stop_time - start_time;
            res = snprintf(buf, AMD_CHAR_BUF_LEN, "%" PRId64 "", total_time);
            if (res < 0 || res > AMD_CHAR_BUF_LEN - 1) {
                switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(fs_s), SWITCH_LOG_ERROR, "Total time truncated [%s], [%d] attempted!\n", buf, res);
                switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Total-time", "ERROR (TRUNCATED)");
            }
            switch_event_add_header_string(event, SWITCH_STACK_BOTTOM, "Total-time", buf);
            break;

        default:
            switch_event_destroy(&event);
            return;
    }

    if ((switch_event_dup(&event_copy, event)) != SWITCH_STATUS_SUCCESS) {
        return;
    }

    switch_core_session_queue_event(fs_s, &event);
    switch_event_fire(&event_copy);
    return;
}

SWITCH_MODULE_LOAD_FUNCTION(mod_amd_load)
{
        switch_application_interface_t *app_interface;

        *module_interface = switch_loadable_module_create_module_interface(pool, modname);

	if (amd_register_all_events() != SWITCH_STATUS_SUCCESS) {
        switch_log_printf(SWITCH_CHANNEL_LOG, SWITCH_LOG_ERROR, "Couldn't register avmd events!\n");
        return SWITCH_STATUS_TERM;
    	}

    	if ((switch_event_bind(modname, SWITCH_EVENT_RELOADXML, NULL, amd_reloadxml_event_handler, NULL) != SWITCH_STATUS_SUCCESS)) {
        switch_log_printf(SWITCH_CHANNEL_LOG, SWITCH_LOG_ERROR, "Couldn't bind our reloadxml handler! Module will not react to changes made in XML configuration\n");
        /* Not so severe to prevent further loading, well - it depends, anyway */
    	}


        do_config(SWITCH_FALSE);

        SWITCH_ADD_APP(
                app_interface,
                "amd",
                "Voice activity detection (blocking)",
                "Asterisk's AMD (Blocking)",
                amd_start_function,
                NULL,
                SAF_NONE);

        return SWITCH_STATUS_SUCCESS;
}

SWITCH_MODULE_SHUTDOWN_FUNCTION(mod_amd_shutdown)
{
	switch_xml_config_cleanup(instructions);

	amd_unregister_all_events();

	return SWITCH_STATUS_SUCCESS;
}

typedef enum {
	SILENCE,
	VOICED
} amd_frame_classifier;

typedef enum {
	VAD_STATE_IN_WORD,
	VAD_STATE_IN_SILENCE,
} amd_vad_state_t;

typedef struct {
	const switch_core_session_t *session;
	switch_channel_t *channel;
	amd_vad_state_t state;
	uint32_t frame_ms;

	uint32_t silence_duration;
	uint32_t voice_duration;
	uint32_t words;

	uint32_t in_initial_silence:1;
	uint32_t in_greeting:1;
} amd_vad_t;

static void amd_reloadxml_event_handler(switch_event_t *event) {
    //amd_load_xml_configuration(amd_globals.mutex);
}

static amd_frame_classifier classify_frame(const switch_frame_t *f, const switch_codec_implementation_t *codec)
{
	int16_t *audio = f->data;
	uint32_t score, count, j;
	double energy;
	int divisor;

	divisor = codec->actual_samples_per_second / 8000;

	for (energy = 0, j = 0, count = 0; count < f->samples; count++) {
		energy += abs(audio[j++]);
		j += codec->number_of_channels;
	}

	score = (uint32_t) (energy / (f->samples / divisor));

	if (score >= globals.silence_threshold) {
		return VOICED;
	}

	return SILENCE;
}

static switch_bool_t amd_handle_silence_frame(amd_vad_t *vad, const switch_frame_t *f)
{
	vad->silence_duration += vad->frame_ms;

	if (vad->silence_duration >= globals.between_words_silence) {
		if (vad->state != VAD_STATE_IN_SILENCE) {
			switch_log_printf(
				SWITCH_CHANNEL_SESSION_LOG(vad->session),
				SWITCH_LOG_DEBUG,
				"AMD: Changed state to VAD_STATE_IN_SILENCE\n");
		}

		vad->state = VAD_STATE_IN_SILENCE;
		vad->voice_duration = 0;
	}

	if (vad->in_initial_silence && vad->silence_duration >= globals.initial_silence) {
		switch_log_printf(
			SWITCH_CHANNEL_SESSION_LOG(vad->session),
			SWITCH_LOG_DEBUG,
			"AMD: MACHINE (silence_duration: %d, initial_silence: %d)\n",
			vad->silence_duration,
			globals.initial_silence);

		switch_channel_set_variable(vad->channel, "amd_result", "MACHINE");
		switch_channel_set_variable(vad->channel, "amd_cause", "INITIALSILENCE");
		return SWITCH_TRUE;
	}

	if (vad->silence_duration >= globals.after_greeting_silence && vad->in_greeting) {
		switch_log_printf(
			SWITCH_CHANNEL_SESSION_LOG(vad->session),
			SWITCH_LOG_DEBUG,
			"AMD: HUMAN (silence_duration: %d, after_greeting_silence: %d)\n",
			vad->silence_duration,
			globals.after_greeting_silence);

		switch_channel_set_variable(vad->channel, "amd_result", "HUMAN");
		switch_channel_set_variable(vad->channel, "amd_cause", "HUMAN");
		return SWITCH_TRUE;
	}

	return SWITCH_FALSE;
}

static switch_bool_t amd_handle_voiced_frame(amd_vad_t *vad, const switch_frame_t *f)
{
	vad->voice_duration += vad->frame_ms;

	if (vad->voice_duration >= globals.minimum_word_length && vad->state == VAD_STATE_IN_SILENCE) {
		vad->words++;

		switch_log_printf(
			SWITCH_CHANNEL_SESSION_LOG(vad->session),
			SWITCH_LOG_DEBUG,
			"AMD: Word detected (words: %d)\n",
			vad->words);

		vad->state = VAD_STATE_IN_WORD;
	}

	if (vad->voice_duration >= globals.maximum_word_length) {
		switch_log_printf(
			SWITCH_CHANNEL_SESSION_LOG(vad->session),
			SWITCH_LOG_DEBUG,
			"AMD: MACHINE (voice_duration: %d, maximum_word_length: %d)\n",
			vad->voice_duration,
			globals.maximum_word_length);

		switch_channel_set_variable(vad->channel, "amd_result", "MACHINE");
		switch_channel_set_variable(vad->channel, "amd_cause", "MAXWORDLENGTH");
		return SWITCH_TRUE;
	}

	if (vad->words >= globals.maximum_number_of_words) {
		switch_log_printf(
			SWITCH_CHANNEL_SESSION_LOG(vad->session),
			SWITCH_LOG_DEBUG,
			"AMD: MACHINE (words: %d, maximum_number_of_words: %d)\n",
			vad->words,
			globals.maximum_number_of_words);

		switch_channel_set_variable(vad->channel, "amd_result", "MACHINE");
		switch_channel_set_variable(vad->channel, "amd_cause", "MAXWORDS");
		return SWITCH_TRUE;
	}

	if (vad->in_greeting && vad->voice_duration >= globals.greeting) {
		switch_log_printf(
			SWITCH_CHANNEL_SESSION_LOG(vad->session),
			SWITCH_LOG_DEBUG,
			"AMD: MACHINE (voice_duration: %d, greeting: %d)\n",
			vad->voice_duration,
			globals.greeting);

		switch_channel_set_variable(vad->channel, "amd_result", "MACHINE");
		switch_channel_set_variable(vad->channel, "amd_cause", "LONGGREETING");
		return SWITCH_TRUE;
	}

	if (vad->voice_duration >= globals.minimum_word_length) {
		if (vad->silence_duration) {
			switch_log_printf(
				SWITCH_CHANNEL_SESSION_LOG(vad->session),
				SWITCH_LOG_DEBUG,
				"AMD: Detected Talk, previous silence duration: %dms\n",
				vad->silence_duration);
		}

		vad->silence_duration = 0;
	}

	if (vad->voice_duration >= globals.minimum_word_length && !vad->in_greeting) {
		if (vad->silence_duration) {
			switch_log_printf(
				SWITCH_CHANNEL_SESSION_LOG(vad->session),
				SWITCH_LOG_DEBUG,
				"AMD: Before Greeting Time (silence_duration: %d, voice_duration: %d)\n",
				vad->silence_duration,
				vad->voice_duration);
		}

		vad->in_initial_silence = 0;
		vad->in_greeting = 1;
	}

	return SWITCH_FALSE;
}

SWITCH_STANDARD_APP(amd_start_function)
{
	switch_channel_t *channel = switch_core_session_get_channel(session);
	switch_media_bug_t  *bug = NULL;
	switch_codec_t raw_codec = { 0 };
	switch_codec_implementation_t read_impl = { 0 };
	switch_frame_t *read_frame;
	switch_status_t status;
	uint32_t timeout_ms = globals.total_analysis_time;
	int32_t sample_count_limit;
	switch_bool_t complete = SWITCH_FALSE;

	amd_vad_t vad = { 0 };

	if (!session) {
		return;
	}

	vad.channel = channel;
	vad.session = session;
	vad.state = VAD_STATE_IN_WORD;
	vad.silence_duration = 0;
	vad.voice_duration = 0;
	vad.frame_ms = 0;
	vad.in_initial_silence = 1;
	vad.in_greeting = 0;
	vad.words = 0;

	switch_core_session_get_read_impl(session, &read_impl);

	if (timeout_ms) {
		sample_count_limit = (read_impl.actual_samples_per_second / 1000) * timeout_ms;
	}

	/*
	 * We are creating a new L16 (raw 16-bit samples) codec for the read end
	 * of our channel.  We'll use this to process the audio coming off of the
	 * channel so that we always know what we are dealing with.
	 */
	status = switch_core_codec_init(
		&raw_codec,
		"L16",
		NULL,
		NULL,
		read_impl.actual_samples_per_second,
		read_impl.microseconds_per_packet / 1000,
		1,
		SWITCH_CODEC_FLAG_ENCODE | SWITCH_CODEC_FLAG_DECODE,
		NULL,
		switch_core_session_get_pool(session));

	if (status != SWITCH_STATUS_SUCCESS) {
		switch_log_printf(
			SWITCH_CHANNEL_SESSION_LOG(session),
			SWITCH_LOG_ERROR,
			"Unable to initialize L16 (raw) codec.\n");
		return;
	}

	switch_core_session_set_read_codec(session, &raw_codec);

	while (switch_channel_ready(channel)) {
		status = switch_core_session_read_frame(session, &read_frame, SWITCH_IO_FLAG_NONE, 0);

		if (!SWITCH_READ_ACCEPTABLE(status)) {
			break;
		}

		if (read_frame->samples == 0) {
			continue;
		}

		vad.frame_ms = 1000 / (read_impl.actual_samples_per_second / read_frame->samples);

		if (sample_count_limit) {
			sample_count_limit -= raw_codec.implementation->samples_per_packet;
			if (sample_count_limit <= 0) {
				switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(session), SWITCH_LOG_DEBUG, "AMD: Timeout\n");

				switch_channel_set_variable(channel, "amd_result", "NOTSURE");
				switch_channel_set_variable(channel, "amd_cause", "TOOLONG");
				break;
			}
		}

		switch (classify_frame(read_frame, &read_impl)) {
		case SILENCE:
			switch_log_printf(
				SWITCH_CHANNEL_SESSION_LOG(session),
				SWITCH_LOG_DEBUG,
				"AMD: Silence\n");

			if (amd_handle_silence_frame(&vad, read_frame)) {
				complete = SWITCH_TRUE;
				amd_fire_event(AMD_EVENT_MACHINE, session, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0);

			}
			break;
		case VOICED:
		default:
			switch_log_printf(
				SWITCH_CHANNEL_SESSION_LOG(session),
				SWITCH_LOG_DEBUG,
				"AMD: Voiced\n");

			if (amd_handle_voiced_frame(&vad, read_frame)) {
				complete = SWITCH_TRUE;
				amd_fire_event(AMD_EVENT_MACHINE, session, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0);
			}
			break;
		}

		if (complete) {
			break;
		}
	}

	bug = (switch_media_bug_t *) switch_channel_get_private(channel, "_avmd_"); /* Is this channel already set? */
    if (bug != NULL) { /* We have already started */
        switch_log_printf(SWITCH_CHANNEL_SESSION_LOG(session), SWITCH_LOG_ERROR, "Avmd already started!\n");
        return;
    }


	 switch_channel_set_private(channel, "_amd_", bug); /* Set the avmd tag to detect an existing avmd media bug */
	 amd_fire_event(AMD_EVENT_SESSION_START, session, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0);

	switch_core_session_reset(session, SWITCH_FALSE, SWITCH_TRUE);
	switch_core_codec_destroy(&raw_codec);
}
