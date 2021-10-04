# pylint: disable=line-too-long,missing-docstring,invalid-name,broad-except,function-redefined,superfluous-parens
import base64
import json
import re
import sys
from os import chmod, remove
from os.path import isfile
from shutil import copyfile
from subprocess import PIPE, Popen, check_output, run

import requests
# from behave import step
from requests_kerberos import HTTPKerberosAuth
from behave import given, when, then

URL = "https://goermis.cern.ch/p/api/v1/alias/"
# url = "https://aiermis.cern.ch/p/api/v1/alias/"
example_alias_name = "test-alias-behave05.cern.ch"
node = "test1.cern.ch"
alarm = "minimum:lb-experts@cern.ch:1"

KERBEROS_FILENAME = ""

# Tip: When running tests outside aiadm (no access to teigi)
# the command should be :
# behave . -D ermists=<base64 password> -D ermistst=<base64 password>


def getacct(context, username):

    try:
        (Output, err) = Popen(['tbag', 'show', '--hg', 'ailbd',
                               username], stdout=PIPE, stderr=PIPE).communicate()
    except:
        password = context.config.userdata[username]
    else:
        password = json.loads(Output)['secret']
    finally:
        print("Got the password of the user %s" % username)
        (Output, err) = Popen(['klist'],
                              stdout=PIPE, stderr=PIPE).communicate()
        print("GOT %s and %s" % (Output, err))
        return base64.b64decode(password)


@given('that we have a valid kerberos ticket of a user in "{n}" egroup')  # pylint: disable=undefined-variable
def step_impl(context, n):  # pylint: disable=unused-argument
    if n == "ermis-lbaas-admins":
        user = "ermistst"
    elif n == "other":
        user = 'ermists'
    else:
        assert False

    kinit = Popen(['kinit', user], stdin=PIPE, stdout=PIPE, stderr=PIPE)
    kinit.communicate(getacct(context, user))
    assert True


@given('that we are "{n}" in the hostgroup')  # pylint: disable=undefined-variable
def step_impl(context, n):
    if n == "admin":
        context.hostgroup = "aiermis"
    elif n == "unauthorized":
        context.hostgroup = "bi"
    else:
        assert False
    assert True


@given('that we have no kerberos ticket')  # pylint: disable=undefined-variable
def step_impl(context):  # pylint: disable=unused-argument
    kdestroy = Popen(['kdestroy'])  # pylint: disable=unused-variable
    assert True


@given('the {object} "{existence}"')  # pylint: disable=undefined-variable
def step_impl(context, object, existence):
    _do_request(context, requests.get, params={
                'alias_name': example_alias_name})
    print(context.response)
    data = context.response.json()
    if object == 'LB alias':
        _check_existence(existence, data["objects"])
    elif object == 'Node':
        allowed = data['objects'][0]['AllowedNodes']
        forbidden = data['objects'][0]['ForbiddenNodes']
        print(allowed)
        print(forbidden)
        _check_existence(existence, allowed + forbidden)
    elif object == 'Alarm':
        _check_existence(existence, data['objects'][0]['alarms'])
    else:
        print("Don't know the object %s" % object)
        assert False


def _check_existence(existence, my_object):
    if existence == "exists":
        assert (my_object != None and my_object != [])
    elif existence == "does not exist":
        assert (my_object == None or my_object == [])
    else:
        print("Don't know what '%s' means", existence)
        assert False


def _do_request(context, my_op, url=URL, data=None, params=None):
    try:
        context.response = my_op(url, data=data, params=params,
                                 headers={'content-type': 'application/json',
                                          'Accept': 'application/json', "WWW-Authenticate": "Negotiate"},
                                 verify='/etc/ssl/certs/CERN-bundle.pem',
                                 auth=HTTPKerberosAuth())
    except Exception as e:
        print(str(e))
        assert False
    assert True


@when('we get the list of aliases')
def step_iml(context):
    _do_request(context, requests.get)


@when('we create an alias')
def step_iml(context):
    payload = {"alias_name": example_alias_name, "best_hosts": 2,
               "external": "external", "metric": "cmsfrontier",
               "polling_interval": 300, "statistics": "none",
               "clusters": "none", "tenant": "", "hostgroup": context.hostgroup}
    _do_request(context, requests.post, data=json.dumps(payload))


@when('we delete an alias')
def step_impl(context):
    _do_request(context, requests.delete, params={
                'alias_name': example_alias_name})


@when('we update an alias')
def step_impl(context):
    _do_request(context, requests.get, params={
                'alias_name': example_alias_name})
    data = context.response.json()
    # resource_uri = data[u'objects'][0][u'resource_uri'].split('/')[5]
    alias_id = data['objects'][0]['alias_id']
    print(alias_id)
    alias_hostgroup = data['objects'][0]['hostgroup']
    payload = {"alias_name": example_alias_name, "best_hosts": 32,
               "external": "external", "metric": "cmsfrontier",
               "polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup": alias_hostgroup}
    _do_request(context, requests.patch, url=URL +
                str(alias_id) + "/", data=json.dumps(payload))


@when('we move an alias')
def step_impl(context):
    _do_request(context, requests.get, params={
                'alias_name': example_alias_name})
    data = context.response.json()
    alias_id = data['objects'][0]['alias_id']
    alias_hostgroup = data['objects'][0]['hostgroup']
    if alias_hostgroup == 'aiermis':
        new_hostgroup = 'bi'
    elif alias_hostgroup == 'bi':
        new_hostgroup = 'aiermis'
    payload = {"alias_name": example_alias_name, "best_hosts": 2,
               "external": "external", "metric": "cmsfrontier",
               "polling_interval": 300, "statistics": "none",
               "clusters": "none", "tenant": "", "hostgroup": new_hostgroup}
    _do_request(context, requests.patch, url=URL +
                str(alias_id) + "/", data=json.dumps(payload))


@when('we {operation} {my_element} in an alias')
def step_impl(context, operation, my_element):
    _do_request(context, requests.get, params={
                'alias_name': example_alias_name})
    data = context.response.json()
    alias_id = data['objects'][0]['alias_id']
    alias_hostgroup = data['objects'][0]['hostgroup']
    payload = {"alias_name": example_alias_name, "best_hosts": 2,
               "external": "external", "metric": "cmsfrontier",
               "polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup": alias_hostgroup}

    if my_element == "a node":
        allowed = {"create": [node], "update": [], "delete": []}
        forbidden = {"create": [], "update": [node], "delete": []}
        payload["AllowedNodes"] = allowed[operation]
        payload["ForbiddenNodes"] = forbidden[operation]
    elif my_element == "an alarm":
        my_ops = {'create': [alarm],
                  'update': ["minimum:lbd-experts@cern.ch:2"],
                  'delete': []}
        payload["alarms"] = my_ops[operation]
    else:
        print("Don't know how to create a %s" % my_element)
        assert False
    print(json.dumps(payload))
    _do_request(context, requests.patch, url=(
        URL + str(alias_id) + "/"), data=json.dumps(payload))


@when('we send a malformed post')
def step_impl(context):
    payload = {"alias_name": "test-alias-behavs_", "best_hosts": "2sd",
               "external": "external", "metric": "cmsfrontier",
               "polling_interval": "ad3f00", "statistics": "none", "clusters": "none", "tenant": "", "hostgroup": ""}
    _do_request(context, requests.post, data=json.dumps(payload))


@then('we get a "{n}" back')  # pylint: disable=undefined-variable
def step_impl(context, n):
    if n == "200":
        assert context.response.status_code == 200, "code %d not expected" % context.response.status_code
    elif n == "202":
        assert context.response.status_code == 202, "code %d not expected" % context.response.status_code
    elif n == "400":
        assert context.response.status_code == 400, "code %d not expected" % context.response.status_code
    elif n == "401":
        assert context.response.status_code == 401, "code %d not expected" % context.response.status_code
    elif n == "409":
        assert context.response.status_code == 409, "code %d not expected" % context.response.status_code
    elif n == "400 or 401":
        # case of unprivileged user returning bad request instead of 401
        assert context.response.status_code == 401 or context.response.status_code == 400
    else:
        assert False


@then('the object should "{req}"')  # pylint: disable=undefined-variable
def step_impl(context, req):  # pylint:disable=too-many-branches,too-many-statements
    # print (context.response.status_code)
    if req == "be created":
        assert context.response.status_code == 201, "code %d not expected" % context.response.status_code
    elif req == "not be created":
        assert context.response.status_code == 400, "code %d not expected" % context.response.status_code
    elif req == "be deleted":
        assert context.response.status_code == 200, "code %d not expected" % context.response.status_code
    else:
        _do_request(context, requests.get, params={
                    'alias_name': example_alias_name})
        data = context.response.json()
        if req == "be updated":
            assert data[u'objects'][0][u'best_hosts'] == 32
        elif req == "have node":
            nodename = data[u'objects'][0][u'AllowedNodes'][0].split(":")[0]
            assert nodename == 'test1.cern.ch'
        elif req == "have updated nodes":
            nodename = data[u'objects'][0][u'ForbiddenNodes'][0].split(":")[0]
            assert nodename == 'test1.cern.ch'
        elif req == "not have node":
            assert data[u'objects'][0][u'ForbiddenNodes'] == []

        elif req == "have alarm":
            r_alarm = data[u'objects'][0][u'alarms'][0]
            # Keep only the first 3 parts of the alarm
            alarm_trunc = ":".join(r_alarm.split(":", 3)[:3])
            assert alarm_trunc == alarm
        elif req == "not have alarm":
            data[u'objects'][0][u'alarms'] = []
        elif req == "have updated alarm":
            r_alarm = data[u'objects'][0][u'alarms'][0]
            # Keep only the first 3 parts of the alarm
            alarm_trunc = ":".join(r_alarm.split(":", 3)[:3])
            assert alarm_trunc == "minimum:lbd-experts@cern.ch:2"
        else:
            assert False
