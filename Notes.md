- Remove all finalizers 
```
kubectl patch pod nginx --type='merge' -p '{"metadata":{"finalizers":null}}'
```
