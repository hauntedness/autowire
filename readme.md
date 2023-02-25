```plantuml
@startuml
package context {
    note top of beans: contains all the objects
    database beans{
        (interface) 
        (struct) 
        (buildin) 
        (pointer)
    }
    note top of packages: store as hashmap in context
    package packages {
        package injectors {
            package providers {
                package input_and_output {
                     database beans_set {
                        note right of beans_set: doesnt include buildin as too complicated to handle
                     }                    
                 }
            }
        }
    }   
    (interface) ..> beans_set
    (struct)    ..> beans_set
    (pointer)   ..> beans_set
}
@enduml
```