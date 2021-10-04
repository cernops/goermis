import requests
from steps.ermis_tests import _do_request, example_alias_name


def before_all(context):
    print("Let's delete the alias (in case it exists)")
    _do_request(context, requests.delete, params={
                'alias_name': example_alias_name})
