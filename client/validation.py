import sys
import const

def check_arguments_validation():
    if len(sys.argv) != 3:
        print("Please enter 2 ip's as arguments: client.py [router_ip] [ip_of_destination]")
        return False
    return validate_ip_address(sys.argv[const.ROUTER_IP_IDX]) and validate_ip_address(sys.argv[const.DESTINATION__IP_IDX])

def validate_ip_address(address):
    parts = address.split(".")

    if len(parts) != 4:
        print("IP address {} is not valid".format(address))
        return False

    for part in parts:
        if not part.isnumeric():
            print("IP address {} is not valid".format(address))
            return False

        if int(part) < 0 or int(part) > 255:
            print("IP address {} is not valid".format(address))
            return False
 
    print("IP address {} is valid".format(address))
    return True 