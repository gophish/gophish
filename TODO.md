# TODO
- ~~Set up separate development / production config~~
- ~~Set up production certificates~~

## Email server things
- Integrate golang email library - https://github.com/toorop/tmail
- Enable/disable STARTTLS for emails
- Enable/disable DKIM for emails
- (Maybe) Enable/disable PGP signature for emails
- Get email certificate for DKIM

## Campaign set up and operations
- Coordinate / test batch user upload
- Create an anonymized identifier field for all targets
- Create a unique tracking ID for each user
- Remove Sending Profiles
- Anonymous data export for NSRG
- Anonymize debug logs so that we can see what is going on in case of failure

## Email template 
- Embed tracking image and track email client User-Agent
- Embed unique link and track user link clicking (and browser User-Agent)
- Record email transmission status (success, fail, etc)

## Phishing site
- Make the phishing website work 
- Track phishing page interaction, links clicked, time spent, etc.
- Track credential entry (user / pass) actions, but not data

## Debrief / education / survey
- Allow opting out of study
- Track education page interaction
- Track follow-up survey through SurveyMonkey API
- Generate survey participation prize winners for TS

## Sysadmin stuff
- Get everything running in production on AWS
	- postgresql
- Seed data for development - campaign, email list, etc.

## Code tidiness / procrasti-work things
- Add more useful campaign tracking features for campaign page or a GUI for visualizing the campaign results
- Figure out why the build now fails on golang v1.5 and v1.6 (investigate commit e5665ca36e05...)


