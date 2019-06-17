"bools" = [true, false, true]
"empty" = {}
"intbools" = [1, 0, 1]
"orphan" = ["red", "blue", "orange"]
"parent1" = {
  "child1" = {
    "empty" = {}
    "grandchild1" = {
      "ids" = [1, 2, 3]
      "on" = true
    }
    "name" = "child1"
    "type" = "hcl"
  }
  "id" = 1234
  "name" = "parent1"
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
"type" = "hcl"
"time" = "2019-01-01"
