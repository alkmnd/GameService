package models

import "time"

type CreateMeetingRequest struct {
	Topic     string   `json:"topic"`
	Type      int      `json:"type"`
	StartTime string   `json:"start_time"`
	Duration  int      `json:"duration"`
	Timezone  string   `json:"timezone"`
	Password  string   `json:"password"`
	Agenda    string   `json:"agenda"`
	Settings  Settings `json:"settings"`
}

type Settings struct {
	HostVideo        bool   `json:"host_video"`
	ParticipantVideo bool   `json:"participant_video"`
	CnMeeting        bool   `json:"cn_meeting"`
	InMeeting        bool   `json:"in_meeting"`
	JoinBeforeHost   bool   `json:"join_before_host"`
	MuteUponEntry    bool   `json:"mute_upon_entry"`
	Watermark        bool   `json:"watermark"`
	UsePmi           bool   `json:"use_pmi"`
	ApprovalType     int    `json:"approval_type"`
	Audio            string `json:"audio"`
	AutoRecording    string `json:"auto_recording"`
}
type CreateMeetingResponse struct {
	AssistantId       string    `json:"assistant_id"`
	HostEmail         string    `json:"host_email"`
	Id                int64     `json:"id"`
	RegistrationUrl   string    `json:"registration_url"`
	Agenda            string    `json:"agenda"`
	CreatedAt         time.Time `json:"created_at"`
	Duration          int       `json:"duration"`
	EncryptedPassword string    `json:"encrypted_password"`
	PstnPassword      string    `json:"pstn_password"`
	H323Password      string    `json:"h323_password"`
	JoinUrl           string    `json:"join_url"`
	ChatJoinUrl       string    `json:"chat_join_url"`
	Occurrences       []struct {
		Duration     int       `json:"duration"`
		OccurrenceId string    `json:"occurrence_id"`
		StartTime    time.Time `json:"start_time"`
		Status       string    `json:"status"`
	} `json:"occurrences"`
	Password    string `json:"password"`
	Pmi         string `json:"pmi"`
	PreSchedule bool   `json:"pre_schedule"`
	Recurrence  struct {
		EndDateTime    time.Time `json:"end_date_time"`
		EndTimes       int       `json:"end_times"`
		MonthlyDay     int       `json:"monthly_day"`
		MonthlyWeek    int       `json:"monthly_week"`
		MonthlyWeekDay int       `json:"monthly_week_day"`
		RepeatInterval int       `json:"repeat_interval"`
		Type           int       `json:"type"`
		WeeklyDays     string    `json:"weekly_days"`
	} `json:"recurrence"`
	Settings struct {
		AllowMultipleDevices               bool   `json:"allow_multiple_devices"`
		AlternativeHosts                   string `json:"alternative_hosts"`
		AlternativeHostsEmailNotification  bool   `json:"alternative_hosts_email_notification"`
		AlternativeHostUpdatePolls         bool   `json:"alternative_host_update_polls"`
		ApprovalType                       int    `json:"approval_type"`
		ApprovedOrDeniedCountriesOrRegions struct {
			ApprovedList []string `json:"approved_list"`
			DeniedList   []string `json:"denied_list"`
			Enable       bool     `json:"enable"`
			Method       string   `json:"method"`
		} `json:"approved_or_denied_countries_or_regions"`
		Audio                   string `json:"audio"`
		AudioConferenceInfo     string `json:"audio_conference_info"`
		AuthenticationDomains   string `json:"authentication_domains"`
		AuthenticationException []struct {
			Email   string `json:"email"`
			Name    string `json:"name"`
			JoinUrl string `json:"join_url"`
		} `json:"authentication_exception"`
		AuthenticationName   string `json:"authentication_name"`
		AuthenticationOption string `json:"authentication_option"`
		AutoRecording        string `json:"auto_recording"`
		BreakoutRoom         struct {
			Enable bool `json:"enable"`
			Rooms  []struct {
				Name         string   `json:"name"`
				Participants []string `json:"participants"`
			} `json:"rooms"`
		} `json:"breakout_room"`
		CalendarType      int    `json:"calendar_type"`
		CloseRegistration bool   `json:"close_registration"`
		ContactEmail      string `json:"contact_email"`
		ContactName       string `json:"contact_name"`
		CustomKeys        []struct {
			Key   string `json:"key"`
			Value string `json:"value"`
		} `json:"custom_keys"`
		EmailNotification     bool     `json:"email_notification"`
		EncryptionType        string   `json:"encryption_type"`
		FocusMode             bool     `json:"focus_mode"`
		GlobalDialInCountries []string `json:"global_dial_in_countries"`
		GlobalDialInNumbers   []struct {
			City        string `json:"city"`
			Country     string `json:"country"`
			CountryName string `json:"country_name"`
			Number      string `json:"number"`
			Type        string `json:"type"`
		} `json:"global_dial_in_numbers"`
		HostVideo              bool `json:"host_video"`
		JbhTime                int  `json:"jbh_time"`
		JoinBeforeHost         bool `json:"join_before_host"`
		LanguageInterpretation struct {
			Enable       bool `json:"enable"`
			Interpreters []struct {
				Email     string `json:"email"`
				Languages string `json:"languages"`
			} `json:"interpreters"`
		} `json:"language_interpretation"`
		SignLanguageInterpretation struct {
			Enable       bool `json:"enable"`
			Interpreters []struct {
				Email        string `json:"email"`
				SignLanguage string `json:"sign_language"`
			} `json:"interpreters"`
		} `json:"sign_language_interpretation"`
		MeetingAuthentication        bool `json:"meeting_authentication"`
		MuteUponEntry                bool `json:"mute_upon_entry"`
		ParticipantVideo             bool `json:"participant_video"`
		PrivateMeeting               bool `json:"private_meeting"`
		RegistrantsConfirmationEmail bool `json:"registrants_confirmation_email"`
		RegistrantsEmailNotification bool `json:"registrants_email_notification"`
		RegistrationType             int  `json:"registration_type"`
		ShowShareButton              bool `json:"show_share_button"`
		UsePmi                       bool `json:"use_pmi"`
		WaitingRoom                  bool `json:"waiting_room"`
		Watermark                    bool `json:"watermark"`
		HostSaveVideoOrder           bool `json:"host_save_video_order"`
		InternalMeeting              bool `json:"internal_meeting"`
		MeetingInvitees              []struct {
			Email string `json:"email"`
		} `json:"meeting_invitees"`
		ContinuousMeetingChat struct {
			Enable                      bool   `json:"enable"`
			AutoAddInvitedExternalUsers bool   `json:"auto_add_invited_external_users"`
			ChannelId                   string `json:"channel_id"`
		} `json:"continuous_meeting_chat"`
		ParticipantFocusedMeeting bool `json:"participant_focused_meeting"`
		PushChangeToCalendar      bool `json:"push_change_to_calendar"`
		Resources                 []struct {
			ResourceType    string `json:"resource_type"`
			ResourceId      string `json:"resource_id"`
			PermissionLevel string `json:"permission_level"`
		} `json:"resources"`
		AutoStartMeetingSummary       bool `json:"auto_start_meeting_summary"`
		AutoStartAiCompanionQuestions bool `json:"auto_start_ai_companion_questions"`
	} `json:"settings"`
	StartTime      time.Time `json:"start_time"`
	StartUrl       string    `json:"start_url"`
	Timezone       string    `json:"timezone"`
	Topic          string    `json:"topic"`
	TrackingFields []struct {
		Field   string `json:"field"`
		Value   string `json:"value"`
		Visible bool   `json:"visible"`
	} `json:"tracking_fields"`
	Type           int    `json:"type"`
	DynamicHostKey string `json:"dynamic_host_key"`
}

type AuthData struct {
	AccessToken string `json:"access_token"`
}
