"bools" = [true, false, true]
"empty" = {}
"intbools" = [1, 0, 1]
"orphan" = ["red", "blue", "orange"]
"parent1" = {
  "boolmap" = {
    "notok3" = false
    "ok1" = true
    "ok2" = true
  }
  "child1" = {
    "empty" = {}
    "grandchild1" = {
      "ids" = [1, 2, 3]
      "on" = true
    }
    "name" = "child1"
    "type" = "hcl"
  }
  "floatmap" = {
    "key1" = 1.1
    "key2" = 1.2
    "key3" = 1.3
  }
  "id" = 1234
  "intmap" = {
    "key1" = 1
    "key2" = 1
    "key3" = 1
  }
  "name" = "parent1"
  "strmap" = {
    "key1" = "val1"
    "key2" = "val2"
    "key3" = "val3"
  }
  "strsmap" = {
    "key1" = ["val1", "val2", "val3"]
    "key2" = ["val4", "val5"]
  }
}
"parent2" = {
  "child2" = {
    "empty" = {}
    "grandchild2" = {
      "ids" = [4, 5, 6]
      "on" = true
    }
    "name" = "child2"
  }
  "id" = 5678
  "name" = "parent2"
}
"strbool" = "1"
"strbools" = ["1", "t", "f"]
"time" = "2019-01-01"
"duration" = "3s"
"negative_int" = -1234
"type" = "hcl"
