# ovhwrapper

## Usage

```
# ovhcon -h
NAME:
   ovhcon - cli tool for the ovh api

USAGE:
   ovhcon <command> [subcommand] [options]

VERSION:
   v0.1.0

COMMANDS:
   list, l            list servicelines and/or clusters
   status, s          show status of a serviceline or cluster
   describe, d        show details of a serviceline and or cluster(s)
   update, u          update k8s cluster
   kubeconfig, kc     kubernetes client configuration
   credentials, cred  shows the credentials used for api access
   logout, o          revoke consumer key, next time the command will be run it will create a new consumer key
   help, h            Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```
Standardschalter sind [-s serviceline], [-c cluster] und [-a] fuer alle, wobei serviceline und cluster mit ihrer ID, ihrem 
Namen oder der Kurzform des Namens angegeben werden koennen.

Hilfe zu den einzelnen Funktionen koennen mit ovhcon <command> -h angezeigt werden.

### list
```
NAME:
   ovhcon list - list servicelines and/or clusters

USAGE:
   ovhcon list [command [command options]] 

OPTIONS:
   --all, -a                      list all servicelines and clusters (default: false)
   --serviceline value, -s value  list clusters of given serviceline
   --help, -h                     show help (default: false)
```

Zeigt eine Liste der servicelines, einer Serviceline und ihrer Cluster (-s) oder alle Servicelines 
und ihre Cluster an (-a).

### status
```
NAME:
   ovhcon status - show status of a serviceline or cluster

USAGE:
   ovhcon status [command [command options]] 

OPTIONS:
   --all, -a                      all servicelines and clusters (default: false)
   --serviceline value, -s value  clusters of a given serviceline
   --cluster value, -c value      specific cluster of a given serviceline
   --help, -h                     show help (default: false)
```

Zeigt den Status eines Clusters (-s und -c), einer Serviceline und ihrer Cluster (nur -s) oder aller 
Servicelines und ihrer Cluster an (-a (ignoriert -s und -c)).

### describe
```
NAME:
   ovhcon describe - show details of a serviceline and or cluster(s)

USAGE:
   ovhcon describe [command [command options]] 

OPTIONS:
   --all, -a                      all servicelines and clusters (default: false)
   --serviceline value, -s value  clusters of a given serviceline
   --cluster value, -c value      specific cluster of a given serviceline
   --output value, -o value       set output format [yaml, json, text]
   --help, -h                     show help (default: false)
```
Zeigt Details der aller Servicelines und ihrer Cluster an (nur -a), einer Serviceline (nur -s) 
und ihrer Cluster (-s und -a) oder eines spezifischen Clusters (-s und -c) und dessen Serviceline an (-s, -c und -a).

Mit --output kann festgelegt werden, ob die Ausgabe im yaml, json oder in Textform (default) ausgegeben werden soll.

### update
```
NAME:
   ovhcon update - update k8s cluster

USAGE:
   ovhcon update [command [command options]] 

OPTIONS:
   --serviceline value, -s value  clusters of a given serviceline
   --cluster value, -c value      specific cluster of a given serviceline
   --force, -f                    force update (default: false)
   --latest, -l                   set strategy to LATEST_PATCH (default is NEXT_MINOR) (default: false)
   --background, -b               if not set the update status will be printed in 1 minute intervals until the cluster is READY again, if background is set the program will exit immediately after starting the upgrade (default: false)
   --help, -h                     show help (default: false)
```
Mit dem Update Kommando wird das Update des Managed Kubernetes Clusters in der ovh Cloud gestartet, der mit Hilfe 
der -s und -c Schalter spezifiziert wurde.

Anschliessend wird der aktuelle Status des Clusters in einem 60 Sekunden Intervall angezeigt, bis der Cluster wieder
READY ist. Optional kann das Statusmonitoring mit --background uebersprungen werden, so dass nur das update angestartet
wird und sich ovhcon danach direkt beendet.

Ist --latest gesetzt wird die Update Strategie von NEXT_MAJOR auf LATEST_PATCH geaendert und direkt die aktuellste
Version installiert ohne die Major Updates dazwischen einzuspielen.

Mit --force kann ein Update forciert werden.

### kubeconfig
```
NAME:
   ovhcon kubeconfig - kubernetes client configuration

USAGE:
   ovhcon kubeconfig [command [command options]] [arguments...]

COMMANDS:
   get, g   get kubeconfig from ovh cloud and save them to file, to certificate files or update entries in a central kubeconfig file
   reset    reset kubeconfig of cluster in the ovh cloud, will redeploy the cluster and reinstall the nodes
   help, h  Shows a list of commands or help for one command

OPTIONS:
   --help, -h  show help (default: false)
```

Das kubeconfig Kommando besteht nur aus 2 Subkommandos:
- _get_ ruft die aktuelle kubeconfig eines Managed Clusters ab
- _reset_ dient dazu die Zertifikate eines Clusters komplett zurueck zu setzen

#### kubeconfig get
```
NAME:
   ovhcon kubeconfig get - get kubeconfig from ovh cloud and save them to file, to certificate files or update entries in a central kubeconfig file

USAGE:
   ovhcon kubeconfig get [command [command options]] 

OPTIONS:
   --all, -a                      all servicelines and clusters (default: false)
   --serviceline value, -s value  serviceline id or name
   --cluster value, -c value      cluster id or name
   --output value, -o value       file, central or certs
   --path value, -p value         output path
   --help, -h                     show help (default: false)
```
kubeconfig get ruft die kubeconfigs entweder aller Servicelines und Cluster (-a), oder eines spezifischen Clusters 
(-s und -c) aus der ovh Cloud ab und speichert diese, mit dem --output Schalter, entweder als einzelne kubeconfig Dateien (default: file), in einer
zentralen central.yaml Datei (central), die alle Konfigurationen enthaelt, gespeichert oder als extrahierte Zertfikate
in einem Unterordner (ca.crt, client.crt und client.key) abgelegt werden. 

Zielordner ist das aktuelle Verzeichnis oder kann mit --path festgelegt werden.

#### kubeconfig reset

``` 
NAME:
   ovhcon kubeconfig reset - reset kubeconfig of cluster in the ovh cloud, will redeploy the cluster and reinstall the nodes

USAGE:
   ovhcon kubeconfig reset [command [command options]] 

OPTIONS:
   --serviceline value, -s value  serviceline id or name
   --cluster value, -c value      cluster id or name
   --background, -b               if not set the cluster status will be printed in 1 minute intervals until the cluster is READY again, if background is set the program will exit immediately after starting the reset (default: false)
   --help, -h                     show help (default: false)
```

kubeconfig reset stoesst ein redeployment des Clusters, der mit -s und -c spezifiziert wurde, und eine Neuinstallation der Nodes an um anschliessend neue
Zertifikate zu haben und damit alte kubeconfigs ungueltig zu machen.

Die Namen und IP Adressen der Nodes bleiben ebenso wie bestehende namespaces und deployments erhalten.

Nach dem anstossen des resets wird der Status gemonitort und alle 60 Sekunden ausgegeben, bis der Cluster wieder im 
READY Status ist. Dieses Monitoring kann mit --background unterbunden werden.

### credentials
```
NAME:
ovhcon credentials - shows the credentials used for api access

USAGE:
ovhcon credentials [command [command options]]

OPTIONS:
--output value, -o value  set output format [yaml, json, text]
--help, -h                show help (default: false)
```

Zeigt die Credentials an, die genutzt werden um Get- und Postzugriffe auf die OVH API durchzufuehren.

Das Output Format kann wie gewohnt mit --output auf text (default), yaml oder json gesteuert werden.

### logout
```
NAME:
ovhcon logout - revoke consumer key, next time the command will be run it will create a new consumer key

USAGE:
ovhcon logout [command [command options]]

OPTIONS:
--help, -h  show help (default: false)
```

Im Gegensatz zum Reader Token ist der Writer Token fuer die OVH API nur fuer begrenzte Zeit (max. 1 Monat) gueltig.
Ausserdem wird dieser dynamisch erzeugt und hat nur Zugriff auf die Cluster, die zum Zeitpunkt der Erzeugung bereits 
vorhanden waren. 

Aus diesem Grund steht die logout Funktion zur Verfuegung, mit deren Hilfe ein noch gueltiger API Key fuer ungueltig 
erklaert werden kann und dieser aus der ovhcon Konfiguration entfernt wird. Ist der Key bereits abgelaufen wird dieser
nur aus der ovhcon Konfiguration entfernt.

Ist beim naechsten Start des ovhcon Tools kein gueltiger Key vorhanden wird ein neuer dynamisch erzeugt und in der 
ovhcon config hinterlegt, dieser muss aber noch ueber die ausgebene url zur OVH WebUI verifiziert werden, bevor er 
genutzt werden kann.
