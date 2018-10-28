# This script prompts the user to enter 
# nN (no) or yY (yes). If yes is provided
# it runs the command in sys.argv[1].
#
# It is meant to be used for confirming
# operations in makefiles.
import sys
from subprocess import call

v = raw_input("Are you sure? (n,Y): ")
if v.lower() == "y":
    call(sys.argv[1].split(" "))