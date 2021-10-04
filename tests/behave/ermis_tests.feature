Feature: Testing API calls for Ermis

  Scenario: test that ermis is up by performing a get request
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
      when we get the list of aliases
      then we get a "200" back

  Scenario: test if unprivileged user can access data with a valid kerberos ticket (works as we currently allow RO access to anyone)
     Given that we have a valid kerberos ticket of a user in "other" egroup
      when we get the list of aliases
      then we get a "200" back

  Scenario: test if unprivileged user can create data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
      when we create an alias
      then we get a "401" back
  
  Scenario: test if hostgroup admin can create data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "does not exist"
      when we create an alias
      then the object should "be created"


   Scenario: test if hostgroup admin can update data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
      when we update an alias
      then we get a "202" back
      and the object should "be updated"

  Scenario: test if hostgroup admin can create node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we create a node in an alias
      then the object should "have node"

  Scenario: test if hostgroup admin can update node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we update a node in an alias
      then the object should "have updated nodes"

   Scenario: test if hostgroup admin can delete node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we delete a node in an alias
      then the object should "not have node"

   Scenario: test if hostgroup admin can create alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Alarm "does not exist"
      when we create an alarm in an alias
      then the object should "have alarm"

  Scenario: test if hostgroup admin can update alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we update an alarm in an alias
      then the object should "have updated alarm"

   Scenario: test if hostgroup admin can delete alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we delete an alarm in an alias
      then the object should "not have alarm"

   Scenario: test if hostgroup admin can delete data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "admin" in the hostgroup
     And the LB alias "exists"
      when we delete an alias
      then the object should "be deleted"

   Scenario: test if privileged user can create data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "does not exist"
      when we create an alias
      then the object should "be created"

   Scenario: test if privileged user can create node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we create a node in an alias
      then the object should "have node"

   Scenario: test if privileged user can update node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we update a node in an alias
      then the object should "have updated nodes"


   Scenario: test if unprivileged can update node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we update a node in an alias
      then we get a "401" back

   Scenario: test if unprivileged can delete node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we delete a node in an alias
      then we get a "401" back

   Scenario: test if privileged user can delete node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "exists"
      when we delete a node in an alias
      then the object should "not have node"


   Scenario: test if privileged user can create alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Alarm "does not exist"
      when we create an alarm in an alias
      then the object should "have alarm"

   Scenario: test if privileged user can update alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we update an alarm in an alias
      then the object should "have updated alarm"


   Scenario: test if unprivileged can update alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we update an alarm in an alias
      then we get a "401" back

   Scenario: test if unprivileged can delete alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we delete an alarm in an alias
      then we get a "401" back

   Scenario: test if privileged user can delete alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we delete an alarm in an alias
      then the object should "not have alarm"


  Scenario: test if unprivileged user can delete data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we delete an alias
      then we get a "401" back

  Scenario: test if unprivileged user can update data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we update an alias
      then we get a "401" back

    Scenario: test if unprivileged can create node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we create a node in an alias
      then we get a "401" back




   Scenario: test if unprivileged can create alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
     And the Alarm "does not exist"
      when we create an alarm in an alias
      then we get a "401" back


   Scenario: test if unprivileged user can move alias to a different hostgroup where he is admin
     Given that we have a valid kerberos ticket of a user in "other" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we move an alias
      then we get a "401" back


  Scenario: test if privileged user can create duplicate data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And that we are "unauthorized" in the hostgroup
     And the LB alias "exists"
      when we create an alias
      then we get a "409" back

  Scenario: test if privileged user can update data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
      when we update an alias
      then we get a "202" back
      and the object should "be updated"

   Scenario: test if privileged can create node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Node "does not exist"
      when we create a node in an alias
      then the object should "have node"

   Scenario: test if privileged can update node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Node "exists"
      when we update a node in an alias
      then the object should "have updated nodes"

   Scenario: test if privileged can delete node with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Node "exists"
      when we delete a node in an alias
      then the object should "not have node"


   Scenario: test if privileged can create alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Alarm "does not exist"
      when we create an alarm in an alias
      then the object should "have alarm"

   Scenario: test if privileged can update alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we update an alarm in an alias
      then the object should "have updated alarm"

   Scenario: test if privileged can delete alarm with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
     And the Alarm "exists"
      when we delete an alarm in an alias
      then the object should "not have alarm"


  Scenario: test if privileged user can delete data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
     And the LB alias "exists"
      when we delete an alias
      then the object should "be deleted"



  Scenario: test if a dummy privileged user can create an alias by supplying bogus data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "ermis-lbaas-admins" egroup
      when we send a malformed post
      then we get a "400" back


  Scenario: test if a dummy unprivileged user can create an alias by supplying bogus data with a valid kerberos ticket
     Given that we have a valid kerberos ticket of a user in "other" egroup
      when we send a malformed post
      then we get a "400" back


  Scenario: test if a user with no kerberos ticket can perform a request on Ermis
     Given that we have no kerberos ticket
      when we get the list of aliases
      then we get a "401" back