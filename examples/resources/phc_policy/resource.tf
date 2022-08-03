resource "phc_policy" "my-policy" {
  name = "my-policy"

  rule {
    operation = "readData"

    comparison {
      subject = "user.id"
      type    = "equals"
      value   = "bob"
    }

    comparison {
      subject = "user.groups"
      type    = "includes"
      values  = ["admin", "doctor"] 
    }

    comparison {
      subject = "user.patients"
      type    = "equals"
      target  = "resource.subjects"
    }
  }


  rule {
    operation = "readMaskedData"

    comparison {
      subject = "user.id"
      type    = "equals"
      value   = "john"
    }
  }
}

