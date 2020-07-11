# pylint: disable=line-too-long,missing-docstring,invalid-name,broad-except,function-redefined,superfluous-parens
from subprocess import Popen, PIPE, check_output
import json
import base64
import re


from os import chmod, remove
from os.path import isfile

from shutil import copyfile

import requests

# from behave import step
from requests_kerberos import HTTPKerberosAuth
url = "https://goermis.cern.ch/api/v1/aliases"
#url = "https://aiermis.cern.ch/p/api/v1/alias/"
# url = "https://lxbrf24b05.cern.ch/p/api/v1/alias/"
headers = {'content-type' : 'application/json', 'Accept': 'application/json', "WWW-Authenticate": "Negotiate"}
cafile = '/etc/ssl/certs/CERN-bundle.pem'
example_alias_name = "test-alias-behave05.cern.ch"
node = 'test1.cern.ch'



KERBEROS_FILENAME = ""


def getacct(context, username):
    (Output, err) = Popen(['tbag', 'show', '--hg', 'ailbd', username], stdout=PIPE, stderr=PIPE).communicate()
    if err:
        password = context.config.userdata[username]
        print(password)
    else:
        password = json.loads(Output)['secret']
    print("Got the password of the user %s" % username)
    (Output, err) = Popen(['klist'], stdout=PIPE, stderr=PIPE).communicate()
    print ("GOT %s and %s" % (Output, err))
    return '%s\n' % base64.b64decode(password)


@given('that we have a valid kerberos ticket of a user in "{n}" egroup')  # pylint: disable=undefined-variable
def step_impl(context, n):  # pylint: disable=unused-argument
    if n == "ermis-lbaas-admins":
        kinit = Popen(['kinit', 'ermistst'], stdin=PIPE, stdout=PIPE, stderr=PIPE)
        kinit.stdin.write(getacct(context, "ermistst"))
        kinit.wait()
    elif n == "other":
        kinit = Popen(['kinit', 'ermists'], stdin=PIPE, stdout=PIPE, stderr=PIPE)
        kinit.stdin.write(getacct(context, "ermists"))
        kinit.wait()
    else:
        assert False
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


@given('the LB alias "{existence}"')  # pylint: disable=undefined-variable
def step_impl(context, existence):
    try:
        context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
    except Exception as e:
        print(str(e))
        assert False
    print (context.response)
    data = context.response.json()
    if existence == "exists":
        assert data["objects"] != None
    elif existence == "does not exist":
        assert data["objects"] == None
@given('the Node "{existence}"')#pylint: disable=undefined-variable
def step_impl(context, existence):
    try:
        context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
    except Exception as e:
        print(str(e))
        assert False
    print (context.response)
    data = context.response.json()
    allowed = str(data['objects'][0]['AllowedNodes'])
    forbidden = str(data['objects'][0]['ForbiddenNodes'])
    print (allowed)
    print (forbidden)
    if existence == "exists":
        assert node in allowed or node in forbidden
    elif existence == "does not exist":
        assert node not in allowed and node not in forbidden


@when('we do a "{req}" request')#pylint:disable=undefined-variable
def step_impl(context, req): #pylint:disable=too-many-branches,too-many-statements
    try:
        if req == "get":
            context.response = requests.get(url , headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
        elif req == "update":
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
            # resource_uri = data[u'objects'][0][u'resource_uri'].split('/')[5]
            alias_id = data['objects'][0]['alias_id']
            alias_hostgroup = data['objects'][0]['hostgroup']
            payload = {"alias_name": example_alias_name, "behaviour": "mindless", "best_hosts": 32, \
                    "external": "external", "metric": "minino", \
                    "polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup": alias_hostgroup}
            context.response = requests.patch(url + r'/' + example_alias_name , data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
        elif req == "move":
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
            # resource_uri = data[u'objects'][0][u'resource_uri'].split('/')[5]
            alias_id = data['objects'][0]['alias_id']
            alias_hostgroup = data['objects'][0]['hostgroup']
            if alias_hostgroup == 'aiermis':
                new_hostgroup = 'bi'
            elif alias_hostgroup == 'bi':
                new_hostgroup = 'aiermis'
            payload = {"alias_name": example_alias_name, "behaviour": "mindless", "best_hosts": 2, \
                    "external": "external", "metric": "minino", \
                    "polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup": new_hostgroup}
            context.response = requests.patch(url + r'/' + example_alias_name, data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
        elif req == "post":
            payload = {"alias_name": example_alias_name, "behaviour": "mindless", "best_hosts": 2, \
                	"external": "external", "metric": "minino", \
                	"polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup": context.hostgroup}
            context.response = requests.post(url, data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            print (json.dumps(payload))
            print("WE HAVE DONE THE POST REQUEST")
            print (url)
            print (headers)
            print(context.response)
        elif req == "delete":
            payload = {'alias_name': example_alias_name}
            context.response = requests.delete(url, data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            print("WE HAVE DONE THE DELETE REQUEST")
            print (url)
            print (headers)
            print(context.response)
        elif req == "malformed post":
            payload = {"alias_name": "test-alias-behavs_", "behaviour": "mindless", "best_hosts": "2sd", \
                	"external": "external", "metric": "minino", \
                	"polling_interval": "ad3f00", "statistics": "none", "clusters": "none", "tenant": "", "hostgroup": ""}
            context.response = requests.post(url, data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
        elif req == "create node":
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
            alias_id = data['objects'][0]['alias_id']
            alias_hostgroup = data['objects'][0]['hostgroup']
            payload = {"AllowedNodes":node, "ForbiddenNodes":"", "alias_name": example_alias_name, "behaviour": "mindless", "best_hosts": 2, \
                	"external": "external", "metric": "minino", \
                	"polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup":alias_hostgroup}
            context.response = requests.patch(url + r'/' + example_alias_name, data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
        elif req == "update nodes":
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
            alias_id = data['objects'][0]['alias_id']
            alias_hostgroup = data['objects'][0]['hostgroup']
            payload = {"AllowedNodes":"", "ForbiddenNodes":node, "alias_name": example_alias_name, "behaviour": "mindless", "best_hosts": 2, \
                	"external": "external", "metric": "minino", \
                	"polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup":alias_hostgroup}
            context.response = requests.patch(url + r'/' + example_alias_name , data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
        elif req == "delete node":
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
            alias_id = data['objects'][0]['alias_id']
            alias_hostgroup = data['objects'][0]['hostgroup']
            payload = {"AllowedNodes":"", "ForbiddenNodes":"", "alias_name": example_alias_name, "behaviour": "mindless", "best_hosts": 2, \
                	"external": "external", "metric": "minino", \
                	"polling_interval": 300, "statistics": "none", "clusters": "none", "tenant": "", "hostgroup":alias_hostgroup}
            context.response = requests.patch(url + r'/' + example_alias_name, data=json.dumps(payload), headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
        else:
            assert False
    except Exception as e:  # pylint: disable=broad-except
        print(str(e))
        assert False
    assert True
@then('we get a "{n}" back')  # pylint: disable=undefined-variable
def step_impl(context, n):
    if   n == "200":
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
        assert context.response.status_code == 401 or context.response.status_code == 400  # case of unprivileged user returning bad request instead of 401
    else:
        assert False


@then('the object should "{req}"')  # pylint: disable=undefined-variable
def step_impl(context, req): #pylint:disable=too-many-branches,too-many-statements
    print (context.response.status_code)
    if req == "be created":
        assert context.response.status_code == 201, "code %d not expected" % context.response.status_code
    elif req == "not be created":
        assert context.response.status_code == 400, "code %d not expected" % context.response.status_code
    elif req == "be deleted":
        assert context.response.status_code == 200, "code %d not expected" % context.response.status_code
    elif req == "be updated":
        try:
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
        except Exception as e:
            print (str(e))
            assert False
        assert data[u'objects'][0][u'best_hosts'] == 32
    elif req == "have node":
        try:
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
        except Exception as e:
            print (str(e))
            assert False
        assert data[u'objects'][0][u'AllowedNodes'] == 'test1.cern.ch'
    elif req == "have updated nodes":
        try:
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
        except Exception as e:
            print (str(e))
            assert False
        assert data[u'objects'][0][u'ForbiddenNodes'] == 'test1.cern.ch'
    elif req == "not have node":
        try:
            context.response = requests.get(url + r'/' + example_alias_name, headers=headers, auth=HTTPKerberosAuth(), verify=cafile)
            data = context.response.json()
        except Exception as e:
            print (str(e))
            assert False
        assert data[u'objects'][0][u'ForbiddenNodes'] == ""
    else:
        assert False
@given('that we have a kerberos token')  # pylint: disable=undefined-variable
def step_impl(context):
    klist_output = check_output(["klist"])
    info = re.search(r'Ticket cache: FILE:(\S*)$', klist_output, re.M)
    if info:
        context.kerberos_filename = info.group(1)
        global KERBEROS_FILENAME  # pylint: disable=global-statement
        KERBEROS_FILENAME = context.kerberos_filename
        assert True
    else:
        assert False


@when('we save the token')  # pylint: disable=undefined-variable
def step_impl(context):
    print("SAVING THE TOKEN")
    context.temporary_filename = "/tmp/behave_token"
    print(context.kerberos_filename)
    copyfile(context.kerberos_filename, context.temporary_filename)
    chmod(context.temporary_filename, 0600)
    assert True


@then('token is saved')  # pylint: disable=undefined-variable
def step_impl(context):
    assert isfile(context.temporary_filename)


@given('that we have the saved kerberos token')  # pylint: disable=undefined-variable
def step_impl(context):
    context.temporary_filename = "/tmp/behave_token"
    assert isfile(context.temporary_filename)


@when('we restore the token')  # pylint: disable=undefined-variable
def step_impl(context):
    copyfile(context.temporary_filename, KERBEROS_FILENAME)
    remove(context.temporary_filename)
    chmod(KERBEROS_FILENAME, 0600)
    # and, in case we are in afs, restore as well the afs tokens
    if isfile('/usr/bin/aklog'):
        check_output(['/usr/bin/aklog'])
    assert True


@then('we have a valid token')  # pylint: disable=undefined-variable
def step_impl(context):  # pylint: disable=unused-argument
    assert isfile(KERBEROS_FILENAME)
