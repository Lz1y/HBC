# HBC
### C2 via bind http  
基于[External C2](https://www.cobaltstrike.com/downloads/externalc2spec.pdf) 实现，提供给Cobalt strike使用。效果为无需开启端口的正向连接。  

### 基本结构
![External C2](https://raw.githubusercontent.com/Lz1y/imggo/master/20190701182150.png)  
以上为External C2的基本结构，由三个角色组建而成，其名称和作用分别为：  
  
- Team Server：Cobalt strike的原生服务，接受、处理控制器传达的信息，并且做出相应的回应或者下发指令。    
- 三方Controller：用户自实现的控制器，用于与客户端进行交互，获取客户端传达的信息，并转发给Team Server。  
- 三方Client：用户自实现的客户端，简单理解为一个加载器，与控制器进行交互，接受beacon.dll并执行，通过命名管道与真正的beacon客户端进行交互。  

而此项目增添了一个新的角色，也能理解为是C2 path，也就是webshell  
- webshell：使用控制器控制webshell在目标机器上执行读写文件的操作，从而使得控制器与客户端进行交互。  
  

### 目录结构
│  web.php   //Webshell  
│  
├─channels   //工具类  
│      socket_channel.go  
│  
├─client     //客户端  
│  │  main.go  
│  │  README.md  
│  │  
│  └─invokedll  
│          invokedll.go  
│  
└─controller //控制器  
        main.go  
         
        
