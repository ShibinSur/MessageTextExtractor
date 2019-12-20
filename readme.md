# Get Message Text
Helps get 10K random message text within past 72 hours from Scalyr(non-redacted) and avoiding skip words

* Optional can pass an argument within 1-72 to provide a start window for searching for Message text
* Generate a CSV for the found entries
* Gets text from Scalyr

## TODO
* Error handling 
* Docs

## How to execute
* On Mac : go run cmd.go <back_start_hour>

# OTP Customer finder

* Run go run otp.go <value between 1 to 24>
* Would get list of customer AuthId sendinf OTP messages in one hour duration from value specified (hours behind now)
* If argument given would take default value (last hour)

