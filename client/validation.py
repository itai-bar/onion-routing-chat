import sys
import re
import const

def check_arguments_validation():
    if len(sys.argv) != 3:
        print("Please enter 2 ip's as arguments: client.py [router_ip] [ip_of_destination]")
        return False

    ok = validate_ip_address(sys.argv[const.ROUTER_IP_IDX]) and validate_ip_address(sys.argv[const.DESTINATION__IP_IDX])

    if ok:
        print("IP addresses are valid")
    else:
        print("IP addresses are not valid")

    return ok 

def validate_ip_address(address):
    return re.match(r"^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$", address)