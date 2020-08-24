Feature: Testing API calls for Ermis

  Scenario: Save the kerberos token
     Given that we have a kerberos token
      when we save the token
      then token is saved
        


  Scenario: test that ermis is up by performing a get request
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
      when we do a "get" request
      then we get a "200" back

  Scenario: test if unprivileged user can access data even if he/she has a valid kerberos ticket (works as we currently allow RO access to anyone)
     Given that we have a valid kerberos ticket of a user in "other" egroup
      when we do a "get" request
      then we get a "200" back




  Scenario: test if unprivileged user can create data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
      when we do a "post" request
      then we get a "401" back

  Scenario: test if hostgroup admin can create data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "does not exist"
      when we do a "post" request
      then the object should "be created"

   Scenario: test if hostgroup admin can update data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
      when we do a "update" request
      then we get a "202" back
      and the object should "be updated"

  Scenario: test if hostgroup admin can create node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we do a "create node" request
      then the object should "have node"

  Scenario: test if hostgroup admin can update node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "update nodes" request
      then the object should "have updated nodes"

   Scenario: test if hostgroup admin can delete node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "delete node" request
      then the object should "not have node"

   Scenario: test if hostgroup admin can delete data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
      when we do a "delete" request
      then the object should "be deleted"
   Scenario: test if privileged user can create data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "does not exist"
      when we do a "post" request
      then the object should "be created"

       Scenario: test if privileged user can create node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we do a "create node" request
      then the object should "have node"

       Scenario: test if privileged user can create node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "update nodes" request
      then the object should "have updated nodes"


       Scenario: test if unprivileged can update node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "update nodes" request
      then we get a "401" back

   Scenario: test if unprivileged can delete node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "delete node" request
      then we get a "401" back

       Scenario: test if privileged user can create node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "delete node" request
      then the object should "not have node"



  Scenario: test if unprivileged user can delete data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we do a "delete" request
      then we get a "401" back

      Scenario: test if unprivileged user can update data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we do a "update" request
      then we get a "401" back

    Scenario: test if unprivileged can create node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we do a "create node" request
      then we get a "401" back

  
   
   
   Scenario: test if unprivileged user can move alias to a different hostgroup where he is admin
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we do a "move" request
      then we get a "401" back


Scenario: test if privileged user can create duplicate data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we do a "post" request
      then we get a "409" back

  Scenario: test if privileged user can update data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
      when we do a "update" request
      then we get a "202" back
      and the object should "be updated"

   Scenario: test if privileged can create node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we do a "create node" request
      then the object should "have node"
   
   Scenario: test if privileged can update node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "update nodes" request
      then the object should "have updated nodes"

   Scenario: test if privileged can delete node if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Node "exists"
      when we do a "delete node" request
      then the object should "not have node"

  Scenario: test if privileged user can delete data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
      when we do a "delete" request
      then the object should "be deleted"



  Scenario: test if a dummy privileged user can create an alias by supplying bogus data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
      when we do a "malformed post" request
      then we get a "400" back


  Scenario: test if a dummy unprivileged user can create an alias by supplying bogus data if he/she has a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
      when we do a "malformed post" request
      then we get a "400" back


  Scenario: test if a user with no kerberos ticket can perform a request on Ermis
     Given that we have no kerberos ticket
      when we do a "get" request
      then we get a "401" back
      
  

 Scenario: restore original token
     Given that we have the saved kerberos token
      when we restore the token
      then we have a valid token