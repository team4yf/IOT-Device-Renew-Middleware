# IOT-Device-Renew-Middleware
IOT Device Renew Middleware


[Google Doc](https://docs.google.com/document/d/1kdXLDb_kQuah-iinMXNIq_qNe8Nq0Pyg59AvkRdE6hI/edit#heading=h.mce78li9b3g5)

Project Code: [DRM].

It’s a middleware for iot devices to renew their online status.
Accept to save the timeout command, and always notify the timeout event `^command/timeout` with the command id.

If the command has been feedback, app should call the `drm` to remove the command timeout event.
otherwise, it will notify anyway.

2 important logic: renew, timeup

Renew
1. Device renews itself 10s.
IOT -> DRM -> renew 10s 
2. DRM publishes an event notification right now if the device does not exist.
	DRM -> notify -> App

Timeup
1. DRM will publish an event notification in 10s if the device does not renew again.
DRM -> notify -> App
2. Check device
	App -> DRM -> check -> T/F


We supply 2 event topics
  - `$drm/{{uuid}}/renew`   // renew the device online expired duration .
  - `$drm/{{uuid}}/message` // register a message .
  - `$d2s/{{uuid}}/partner/feedback` // subscribe the feedback
There are 2 event notifications
  - `$drm/{{uuid}}/offline` // notify a message after the device offline .
  - `$drm/{{uuid}}/timeout` // notify a message when the message timtout


#### mqtt client for ubuntu

sudo apt-get install mosquitto-clients

mosquitto_sub -h open.yunplus.io -t "^drm/offline/foo" -u "fpmuser" -P "123123123"

mosquitto_pub -h open.yunplus.io -t "^drm/offline/foo" -m "2" -u "fpmuser" -P "123123123"