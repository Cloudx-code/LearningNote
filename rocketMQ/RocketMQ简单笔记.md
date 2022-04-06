#  一、概述：

## 1.主要用途：

### 1.限流削峰

MQ可以将系统的超量请求暂存其中，以便系统后期可以慢慢进行处理，从而避免了请求的丢失或系统被压垮。

<img src="\RocketMQ简单笔记.assets\image-20220317134534348.png" alt="image-20220317134534348" style="zoom: 33%;" />

### 2.异步解耦

上游系统对下游系统的调用若为同步调用，则会大大降低系统的吞吐量与并发度，且系统耦合度太高。
而异步调用则会解决这些问题。所以两层之间若要实现由同步到异步的转化，一般性做法就是，在这两层间添加一个MQ层。

<img src="\RocketMQ简单笔记.assets\image-20220317141324597.png" alt="image-20220317141324597" style="zoom:33%;" />

### 3.数据收集

分布式系统会产生海量级数据流，如：业务日志、监控数据、用户行为等。针对这些数据流进行实时或批量采集汇总，然后对这些数据流进行大数据分析，这是当前互联网平台的必备技术。通过MQ完成此类数据收集是最好的选择。

## 2.其他内容

大致讲了一些其他MQ产品，以后可能还要关注下Kafka,介绍了一些MQ常用协议，不过RocketMQ和Kafka好像都不遵循这些协议。

# 二、RocketMQ各种细节

## 1.涉及的基本概念

### 消息（Message）

消息是指，消息系统所传输信息的物理载体，生产和消费数据的最小单位，每条消息必须属于一个主题。

### 主题(Topic)

<img src="\RocketMQ简单笔记.assets\image-20220319203245663.png" alt="image-20220319203245663" style="zoom: 33%;" />

Topic表示一类消息的集合（比如上图中狗吃的骨头可以有很多个。）

每个主题包含若干条消息，每条消息只能属于一个主题，是RocketMQ进行消息订阅的基本单位。 

topic:message 1:n 				message:topic 1:1（其中一根狗骨头：狗骨头）

一个生产者可以同时发送多种Topic的消息，而一个消费者只对某种特定的Topic感兴趣，即只可以订阅和消费一种

Topic的消息。 producer:topic 1:n （我给家里不同的动物投的食物，这里举这个例子完全是为了类比上图）		consumer:topic 1:1（每个动物只能吃自己的那份，然而家里狗子啥都吃）

（理解起来有点怪啊，举个例子：我可以给家里的狗为骨头，鱼喂鱼粮，但是狗不能吃鱼粮这种...此处例子怪是因为老师讲的不行，跟我没关系）；

### 标签(Tag)

为消息设置的标签，用于同一主题下区分不同类型的消息。

来自同一业务单元的消息，可以根据不同业务目的在同一主题下设置不同标签。

标签能够有效地保持代码的清晰度和连贯性，并优化RocketMQ提供的查询系统。

消费者可以根据Tag实现对不同子主题的不同消费逻辑，实现更好的扩展性。

**Topic是消息的一级分类，Tag是消息的二级分类。**（这个总结还蛮好的，脑海里想了一下狗子吃的不同狗粮品牌。）

### 队列(Queue)

存储消息的物理实体。一个Topic中可以包含多个Queue，每个Queue中存放的就是该Topic的消息。

一个Topic的Queue也被称为一个Topic中消息的分区（Partition）。

一个Topic的Queue中的消息只能被一个**消费者组**中的一个消费者消费。

一个Queue中的消息**不允许**同一个**消费者组**中的多个消费者同时消费。言外之意就是**允许**不同**消费者组**的一起消费。比如下图的那个老人左边要是再来一群狗，就可以和右边的三只狗一起吃。

<img src="RocketMQ简单笔记.assets\image-20220319204524128.png" alt="image-20220319204524128" style="zoom:33%;" />

在学习参考其它相关资料时，还会看到一个概念：分片（Sharding）。分片不同于分区。在RocketMQ中，分片指的是存放相应Topic的Broker。每个分片中会创建出相应数量的分区，即Queue，每个Queue的大小都是相同的。

<img src="RocketMQ简单笔记.assets\image-20220319204801922.png" alt="image-20220319204801922" style="zoom:33%;" />

### 消息标识(MessageId/Key)

RocketMQ中每个消息拥有唯一的MessageId，且可以携带具有业务标识的Key，以方便对消息的查询。
不过需要注意的是，MessageId有两个：在生产者send()消息时会自动生成一个MessageId（msgId)，
当消息到达Broker后，Broker也会自动生成一个MessageId(offsetMsgId)。msgId、offsetMsgId与key都
称为消息标识。

**msgId**：由producer端生成，其生成规则为：（这个重复概率小）

producerIp + 进程pid + MessageClientIDSetter类的ClassLoader的hashCode + 当前时间 + AutomicInteger自增计数器

**offsetMsgId**：由broker端生成，其生成规则为：（这个重复概率大）

 brokerIp + 物理分区的offset（Queue中的偏移量）

**key**：由用户指定的业务相关的唯一标识

## 2.系统架构

![image-20220326165920938](RocketMQ简单笔记.assets\image-20220326165920938.png)

简单来说，就是分了这四块集群，然后后面来总结下每一块的功能是什么。

### Producer 

消息生产者，负责生产消息。Producer通过MQ的负载均衡模块选择相应的Broker集群队列进行消息投递，投递的过程支持快速失败并且低延迟。

> 例如，业务系统产生的日志写入到MQ的过程，就是消息生产的过程
>
> 再如，电商平台中用户提交的秒杀请求写入到MQ的过程，就是消息生产的过程

RocketMQ中的消息生产者都是以生产者组（Producer Group）的形式出现的。生产者组是同一类生产者的集合，这类Producer发送相同Topic类型的消息。一个生产者组可以同时发送多个主题的消息。（因为一个生成者可以关联多个Topic）

### Consumer

消息消费者，负责消费消息。一个消息消费者会从Broker服务器中获取到消息，并对消息进行相关业务处理。(Broker后面再看)

> 例如，QoS系统从MQ中读取日志，并对日志进行解析处理的过程就是消息消费的过程。
> 再如，电商平台的业务系统从MQ中读取到秒杀请求，并对请求进行处理的过程就是消息消费的过程。

RocketMQ中的消息消费者都是以消费者组（Consumer Group）的形式出现的。消费者组是同一类消费者的集合，这类Consumer消费的是同一个Topic类型的消息。

消费者组使得在消息消费方面，实现以下两个目标变得非常容易。：

**负载均衡**（将一个Topic中的不同的**Queue**平均分配给同一个Consumer Group的不同Consumer，注意，并不是将消息负载均衡）

**容错**（一个Consmer挂了，该Consumer Group中的其它Consumer可以接着消费原Consumer消费的Queue）

两个不太重要的小细节：

（1）消费者组中Consumer的数量应该小于等于订阅Topic的Queue数量。如果超出Queue数量，则多出的Consumer将不能消费消息。（比如下面这个图，Consumer3不能消费任何信息）

<img src="RocketMQ简单笔记.assets\image-20220326171541482.png" alt="image-20220326171541482" style="zoom: 67%;" />

（2）不过，一个Topic类型的消息可以被多个消费者组同时消费。

<img src="RocketMQ简单笔记.assets\image-20220326171819413.png" alt="image-20220326171819413" style="zoom:67%;" />

> 注意，
> 1）消费者组只能消费一个Topic的消息，不能同时消费多个Topic消息
> 2）一个消费者组中的消费者必须订阅完全相同的Topic

### NameServer

#### 大致功能

（先宏观上看下这个模块，然后拆分进去看）

NameServer是一个Broker与Topic路由的注册中心，支持Broker的**动态注册**与**发现**。(相对于用这个NameServer取代了原先Zookeeper的功能模块)

RocketMQ的思想来自于Kafka，而Kafka是依赖了Zookeeper的。所以，在RocketMQ的早期版本，即在MetaQ v1.0与v2.0版本中，也是依赖于Zookeeper的。从MetaQ v3.0，即RocketMQ开始去掉了Zookeeper依赖，使用了自己的NameServer。

**知识点小回顾：**

CAP原则又称CAP定理，指的是在一个[分布式系统](https://so.csdn.net/so/search?q=分布式系统&spm=1001.2101.3001.7020)中，***一致性***（Consistency）、***可用性***（Availability）、***分区容错性***（Partition tolerance）

**原因在于：**

（1）Zookeeper 是CP的，它的一致性影响RocketMQ性能,RocketMQ不需要一致性，貌似他是幂等的

（2）强依赖于第三方，缺乏独立性

（3）搭建成本变高

好了，说到重点了，这个模块主要包括两个功能：

**Broker管理**：接受Broker集群的注册信息并且保存下来作为路由信息的基本数据；提供心跳检测机制，检查Broker是否还存活。

**路由信息管理**：每个NameServer中都保存着Broker集群的整个路由信息和用于客户端查询的队列信息。Producer和Conumser通过

NameServer可以获取整个Broker集群的路由信息，从而进行消息的投递和消费。

下面来分别介绍几个具体功能：

#### 路由注册

NameServer通常也是以集群的方式部署，不过，NameServer是无状态的，即NameServer集群中的**各个节点间是无差异**的（$\textcolor{RoyalBlue}{也就是说每个节点的内容都一样？}$），**各节点间相互不进行信息通讯**。那各节点中的数据是如何进行数据同步的呢？在Broker节点启动时，轮询NameServer列表，与每个NameServer节点建立长连接，发起注册请求。在NameServer内部维护着⼀个Broker列表，用来动态存储Broker的信息。

$\textcolor{RoyalBlue}{由于NameServer之间相互独立，很明显，NameServer是一个AP设计}$

> 注意，这是与其它像zk(zookeeper)、Eureka、Nacos等注册中心不同的地方。
>
> 这种NameServer的无状态方式，有什么优缺点：
>
> 优点：NameServer集群搭建简单，扩容简单。(缩容简单，不能随便扩容)
>
> 缺点：对于Broker，必须明确指出所有NameServer地址。否则未指出的将不会去注册。也正因为如此，NameServer并不能随便扩
>
> 容。因为，若Broker不重新配置，新增的NameServer对于Broker来说是不可见的，其不会向这个NameServer进行注册。

Broker节点为了证明自己是活着的，为了维护与NameServer间的长连接，会将最新的信息以**心跳包**的方式上报给NameServer，每30秒

发送一次心跳。**心跳包中包含 BrokerId、Broker地址(IP+Port)、Broker名称、Broker所属集群名称等等**。NameServer在接收到心跳包后，会更新心跳时间戳，记录这个Broker的最新存活时间。

$\textcolor{RoyalBlue}{关于锁的额外补充}$

>NameServer在处理心跳包的时候，存在多个Broker同时操作一张Broker表，为了防止并发修改Broker表导致不安全，路由注册操作引入了ReadWriteLock读写锁，这个设计亮点允许多个消息生产者并发读，保证了消息发送时的高并发，但是同一时NameServer只能处理一个Broker心跳包，多个心跳包串行处理。这也是读写锁的经典使用场景，即读多写少。

#### 路由剔除

由于Broker关机、宕机或网络抖动等原因，NameServer没有收到Broker的心跳，NameServer可能会将其从Broker列表中剔除。

NameServer中有⼀个定时任务，每隔10秒就会扫描⼀次Broker表，查看每一个Broker的最新心跳时间戳距离当前时间是否超过120秒，

如果超过，则会判定Broker失效，然后将其从Broker列表中剔除。

> 扩展：对于RocketMQ日常运维工作，例如Broker升级，需要停掉Broker的工作。OP(运维工程师)需要怎么做？
> OP需要将Broker的读写权限禁掉。一旦client(Consumer或Producer)向broker发送请求，都会收到broker的NO_PERMISSION响应，然后client会进行对其它Broker的重试。当OP观察到这个Broker没有流量后，再关闭它，实现Broker从NameServer的移除。

#### 路由发现(包含重要知识点)

RocketMQ的路由发现采用的是Pull模型。当Topic路由信息出现变化时，NameServer不会主动推送给客户端，而是客户端定时拉取主题

最新的路由。默认客户端每30秒会拉取一次最新的路由。

>扩展：
>1）Push模型：推送模型。其实时性较好，是一个“发布-订阅”模型，需要维护一个长连接。而长连接的维护是需要资源成本的。该模型适合于的场景：实时性要求较高，Client数量不多，Server数据变化较频繁
>
>2）Pull模型：拉取模型。存在的问题是，实时性较差。
>
>3）Long Polling模型：长轮询模型。其是对Push与Pull模型的整合，充分利用了这两种模型的优势，屏蔽了它们的劣势。
>
>​	  比如nacos中就采用这种方式，每隔一段时间去Pull一下，然后Pull的同时维护一个30s的长连接，之后在断掉。

#### 客户端NameServer选择策略

> 这里的客户端指的是Producer与Consumer

客户端在配置时必须要写上NameServer集群的地址，那么客户端到底连接的是哪个NameServer节点呢？客户端首先会生产一个随机数，然后再与NameServer节点数量取模，此时得到的就是所要连接的节点索引，然后就会进行连接。如果连接失败，则会采用round-robin策略，逐个尝试着去连接其它节点。
首先采用的是**随机策略**进行的选择，失败后采用的是**轮询策略**。

> 扩展：Zookeeper Client是如何选择Zookeeper Server的？
>
> 简单来说就是，经过两次Shuffle，然后选择第一台Zookeeper Server。
> 详细说就是，将配置文件中的zk server地址进行第一次shuffle，然后随机选择一个。这个选择出的一般都是一个hostname。然后获取到该hostname对应的所有ip，再对这些ip进行第二次shuffle，从shuffle过的结果中取第一个server地址进行连接。

$\textcolor{RoyalBlue}{关于NameServer设计层面的额外补充}$

> 在降低NameServer实现复杂性方面，还有一个设计亮点就是NameServer之间是彼此独立无交流的，也就是说NameServer服务器之间在某个时刻的数据并不会完全相同，但是异常重试机制使得这种差异不会造成任何影响。

### Broker

#### 功能介绍

Broker充当着消息中转角色，负责存储消息、转发消息。Broker在RocketMQ系统中负责接收并存储从生产者发送来的消息，同时为消费者的拉取请求作准备。Broker同时也存储着消息相关的元数据，包括消费者组消费进度偏移offset、主题、队列等。

> Kafka 0.8版本之后，offset是存放在Broker中的，之前版本是存放在Zookeeper中的。

#### 模块构成

下图为Broker Server的功能模块示意图。（列出了较为重要的组件）

<img src="RocketMQ简单笔记.assets\image-20220402212829820.png" alt="image-20220402212829820" style="zoom: 67%;" />

$\textcolor{red}{Remoting Module}$：整个Broker的实体，负责处理来自clients端的请求。而这个Broker实体则由以下模块构成。

$\textcolor{red}{Client Manager}$：客户端管理器。负责接收、解析客户端(Producer/Consumer)请求，管理客户端。例如，维护Consumer的Topic订阅信息

$\textcolor{red}{Store Service}$：存储服务。提供方便简单的API接口，处理消息存储到物理硬盘和消息查询功能。HA Service：高可用服务，提供Master Broker 和 Slave Broker之间的数据同步功能。

$\textcolor{red}{Index Service}$：索引服务。根据特定的Message key，对投递到Broker的消息进行索引服务，同时也提供根据Message Key对消息进行快速查询的功能。

#### 集群部署

![image-20220402213357108](\RocketMQ简单笔记.assets\image-20220402213357108.png)



为了增强Broker性能与吞吐量，Broker一般都是以集群形式出现的。各集群节点中可能存放着相同Topic的不同Queue。不过，这里有个问题，如果某Broker节点宕机，如何保证数据不丢失呢？$\textcolor{RoyalBlue}{slave节点的作用}$

其解决方案是，将每个Broker集群节点进行横向扩展，即将Broker节点再建为一个HA（高性能可用）集群，解决单点问题。

Broker节点集群是一个主从集群，即集群中具有Master与Slave两种角色。Master负责处理读写操作请求，Slave负责对Master中的数据进行备份。当Master挂掉了，Slave则会自动切换为Master去工作。所以这个Broker集群是**主备集群**（不是主从集群）。一个Master可以包含多个Slave，但一个Slave只能隶属于一个Master。
Master与Slave 的对应关系是通过指定相同的BrokerName、不同的BrokerId 来确定的。BrokerId为0表示Master，非0表示Slave。每个Broker与NameServer集群中的所有节点建立长连接，定时注册Topic信息到所有NameServer。

### 工作流程

#### 具体流程

<img src="RocketMQ简单笔记.assets\image-20220403004418605.png" alt="image-20220403004418605" style="zoom:50%;" />

1）启动NameServer，NameServer启动后开始监听端口，等待Broker、Producer、Consumer连接。

2）启动Broker时，Broker会与所有的NameServer建立并保持长连接，然后每30秒向NameServer定时发送心跳包。

3）发送消息前，可以先创建Topic，创建Topic时需要指定该Topic要存储在哪些Broker上，当然，在创建Topic时也会将Topic与Broker的关系写入到NameServer中。不过，这步是可选的，也可以在发送消息时自动创建Topic。

4）Producer发送消息，启动时先跟NameServer集群中的其中一台建立长连接，并从NameServer中获取路由信息，即当前发送的Topic消息的Queue与Broker的地址（IP+Port）的映射关系。然后根据算法策略从中选择一个Queue，与队列所在的Broker建立长连接从而向Broker发消息。当然，在获取到路由信息后，Producer会首先将路由信息缓存到本地，再每30秒从NameServer更新一次路由信息。

5）Consumer跟Producer类似，跟其中一台NameServer建立长连接，获取其所订阅Topic的路由信息，然后根据算法策略从路由信息中获取到其所要消费的Queue，然后直接跟Broker建立长连接，开始消费其中的消息。Consumer在获取到路由信息后，同样也会每30秒从NameServer更新一次路由信息。**不过不同于Producer的是，Consumer还会向Broker发送心跳，以确保Broker的存活状态。**

#### Topic的创建模式

手动创建Topic时，有两种模式：

* 集群模式：该模式下创建的Topic在该集群中，所有Broker中的Queue数量是相同的。
* Broker模式：该模式下创建的Topic在该集群中，每个Broker中的Queue数量可以不同。

自动创建Topic时，默认采用的是Broker模式，会为每个Broker默认创建4个Queue。

#### 读/写队列

从物理上来讲，读/写队列是同一个队列。所以，不存在读/写队列数据同步问题。读/写队列是逻辑上进行区分的概念。一般情况下，读/写队列数量是相同的。

>例如，创建Topic时设置的写队列数量为8，读队列数量为4，此时系统会创建8个Queue，分别是0 1 2 3 4 5 6 7。Producer会将消息写入到这8个队列，但Consumer只会消费0 1 2 3这4个队列中的消息，4 5 6 7中的消息是不会被消费到的。
>
>再如，创建Topic时设置的写队列数量为4，读队列数量为8，此时系统会创建8个Queue，分别是0 1 2 3 4 5 6 7。Producer会将消息写入到0 1 2 3 这4个队列，但Consumer会消费0 1 2 3 4 5 6 7这8个队列中的消息，但是4 5 6 7中是没有消息的。此时假设Consumer Group中包含两个Consuer，Consumer1消费0 1 2 3，而Consumer2消费4 5 6 7。但实际情况是，Consumer2是没有消息可消费的。

也就是说，当读/写队列数量设置不同时，总是有问题的。那么，为什么要这样设计呢？

其这样设计的目的是为了，方便Topic的Queue的缩容。

>例如，原来创建的Topic中包含16个Queue，如何能够使其Queue缩容为8个，还不会丢失消息？可以动态修改写队列数量为8，读队列数量不变。此时新的消息只能写入到前8个队列，而消费都消费的却是16个队列中的数据。当发现后8个Queue中的消息消费完毕后，就可以再将读队列数量动态设置为8。整个缩容过程，没有丢失任何消息。

perm用于设置对当前创建Topic的操作权限：2表示只写，4表示只读，6表示读写。

## 3.单机安装与启动

<img src="RocketMQ简单笔记.assets\image-20220404225337587.png" alt="image-20220404225337587" style="zoom: 50%;" />

1.前提条件：好像只用JDK就行了，感觉Maven也要。然后直接安装了解压。

2.限于虚拟机大小，启动前要根据情况修改初始内存等参数。

3.然后启动，启动完了测试后关闭。

## 4.控制台的安装与启动

通过https://github.com/apache/rocketmq-externals/releases链接下载控制台，完了以后改改配置添加Maven依赖后用maven打包再启动（目测类似于一个NameServer）

## 5.集群搭建理论

<img src="RocketMQ简单笔记.assets\image-20220405145247023.png" alt="image-20220405145247023" style="zoom:50%;" />

###  数据复制与刷盘策略

<img src="RocketMQ简单笔记.assets\image-20220405145324670.png" alt="image-20220405145324670" style="zoom:50%;" />

**复制策略**
复制策略是Broker的Master与Slave间的数据同步方式。分为同步复制与异步复制：

- 同步复制：消息写入master后，master会等待slave同步数据成功后才向producer返回成功ACK
- 异步复制：消息写入master后，master立即向producer返回成功ACK，无需等待slave同步数据成功

> 异步复制策略会降低系统的写入延迟，RT(Response Time)变小，提高了系统的吞吐量

**刷盘策略**

刷盘策略指的是broker中消息的**落盘**方式，即消息发送到broker内存后消息持久化到磁盘的方式。分为同步刷盘与异步刷盘：

- 同步刷盘：当消息持久化到broker的磁盘后才算是消息写入成功。
- 异步刷盘：当消息写入到broker的内存后即表示消息写入成功，无需等待消息持久化到磁盘。

> 1）异步刷盘策略会降低系统的写入延迟，RT变小，提高了系统的吞吐量
> 2）消息写入到Broker的内存，一般是写入到了PageCache
> 3）对于异步 刷盘策略，消息会写入到PageCache后立即返回成功ACK。但并不会立即做落盘操作，而是当PageCache到达一定量时会自动进行落盘。

### Broker集群模式

根据Broker集群中各个节点间关系的不同，Broker集群可以分为以下几类：
**单Master**
只有一个broker（其本质上就不能称为集群）。这种方式也只能是在测试时使用，生产环境下不能使用，因为存在单点问题。

**多Master**
broker集群仅由多个master构成，不存在Slave。同一Topic的各个Queue会平均分布在各个master节点上。
优点：配置简单，单个Master宕机或重启维护对应用无影响$\textcolor{RoyalBlue}{（还是有影响的）}$，在磁盘配置为RAID10时，即使机器宕机不可恢复情况下，由于RAID10磁盘非常可靠，消息也不会丢（异步刷盘丢失少量消息，同步刷盘一条不丢），性能最高；
缺点：单台机器宕机期间，这台机器上未被消费的消息在机器恢复之前不可订阅（不可消费），消息实时性会受到影响。

**多Master多Slave模式-异步复制**
broker集群由多个master构成，每个master又配置了多个slave（在配置了RAID磁盘阵列的情况下，一个master一般配置一个slave即可）。master与slave的关系是主备关系，即master负责处理消息的读写请求，而slave仅负责消息的备份与master宕机后的角色切换。

异步复制即前面所讲的**复制策略**中的**异步复制策略**，即消息写入master成功后，master立即向producer返回成功ACK，无需等待slave同步数据成功。

该模式的最大特点之一是，当master宕机后slave能够**自动切换**$\textcolor{RoyalBlue}{（好像不太能自动切换？）}$为master。不过由于slave从master的同步具有短暂的延迟（毫秒级），所以当master宕机后，这种异步复制方式可能会存在少量消息的丢失问题。

> Slave从Master同步的延迟越短，其可能丢失的消息就越少
>
> 对于Master的RAID磁盘阵列，若使用的也是异步复制策略，同样也存在延迟问题，同样也可能会丢失消息。但RAID阵列是微秒级的（因为是由硬盘支持的），所以其丢失的数据量会更少。

**多Master多Slave模式-同步双写**

该模式是多Master多Slave模式的**同步复制**实现。所谓同步双写，指的是消息写入master成功后，master会等待slave同步数据成功后才向producer返回成功ACK，即master与slave都要写入成功后才会返回成功ACK，也即双写。

该模式与**异步复制模式**相比，优点是消息的安全性更高，不存在消息丢失的情况。但单个消息的RT略高，从而导致性能要略低（大约低10%）。
该模式存在一个大的问题：对于目前的版本，Master宕机后，Slave **不会自动切换**$\textcolor{RoyalBlue}{（所以应该咋切换呢？这里也没细说）}$到Master。

**最佳实践**
一般会为Master配置RAID10磁盘阵列，然后再为其配置一个Slave。即利用了RAID10磁盘阵列的高
效、安全性，又解决了可能会影响订阅的问题。$\textcolor{RoyalBlue}{（根据弹幕，大多数公司还是用的多Master多Slave模式）}$

> 1）RAID磁盘阵列的效率要高于Master-Slave集群。因为RAID是硬件支持的。也正因为如此，所以RAID阵列的搭建成本较高。
> 2）多Master+RAID阵列，与多Master多Slave集群的区别是什么？
> **多Master+RAID阵列**，其仅仅可以保证数据不丢失，即不影响消息写入，但其可能会影响到消息的订阅。但其执行效率要远高于多Master多Slave集群
> **多Master多Slave集群**，其不仅可以保证数据不丢失，也不会影响消息写入。其运行效率要低于多Master+RAID阵列

## 6.磁盘阵列RAID（补充）

主要了解了：

`RAID的概念`，即**廉价冗余磁盘阵列**（ Redundant Array of Inexpensive Disks ）。后面磁盘便宜了，RAID 变成了**独立磁盘冗余阵列**（ Redundant Array of Independent Disks ）。但这仅仅是名称的变化，实质内容没有改变。

`RAID的等级`，可以分成几类：主要了解RAID0、RAID1、RAID10、RAID01

> 还有一个JBOD可以拿来对比理解，JBOD：只是简单提供一种扩展存储空间的机制
>
> RAID0 是一种简单的、无数据校验的**数据条带化技术**。RAID0 的性能在所有 RAID 等级中是最高的。
>
> 应用场景：对数据的顺序读写要求不高，对数据的安全性和可靠性要求不高，但对系统性能要求很高的场景。（例如视频通话）
>
> RAID0与JBOD相同点：
> 1）存储容量：都是成员磁盘容量总和
> 2）磁盘利用率，都是100%，即都没有做任何的数据冗余备份
> RAID0与JBOD不同点：
> JBOD：数据是顺序存放的，一个磁盘存满后才会开始存放到下一个磁盘
> RAID：各个磁盘中的数据写入是并行的，是通过数据条带技术写入的。其读写性能是JBOD的n倍(实际上会低一点，这点在课件上有说明。)
>
> RAID1 就是一种**镜像技术**
>
> 应用场景：对顺序读写性能要求较高，或对数据安全性要求较高的场景。

`RAID用到的技术`：镜像技术（用的最多）、数据条带技术、数据校验技术

> 镜像技术提供了非常高的数据安全性，其代价也是非常昂贵的，需要至少双倍的存储空间。高成本限制了镜像的广泛应用，主要应用于至关重要的数据保护，这种场合下的数据丢失可能会造成非常巨大的损失。
>
> 数据条带化技术是一种自动将 I/O操作负载均衡到多个物理磁盘上的技术。更具体地说就是，将一块连续的数据分成很多小部分并把它们分别存储到不同磁盘上。这就能使多个进程可以并发访问数据的多个不同部分，从而获得最大程度上的 I/O 并行能力，极大地提升性能。

`RAID分类`：软RAID、硬RAID、混合RAID（主要区别就是硬件成本的高低以及依赖CPU资源的程度（或者不依赖CPU资源））

## 7.集群搭建实践

大致就是各种配置，实战的话要在多台机器上配置，列几个配置看看吧，感觉工作里不会配这些，这也不是学习的重点。

```properties
### broker-a.properties
# 指定整个broker集群的名称，或者说是RocketMQ集群的名称
brokerClusterName=DefaultCluster
# 指定master-slave集群的名称。一个RocketMQ集群可以包含多个master-slave集群
brokerName=broker-a
# master的brokerId为0
brokerId=0
# 指定删除消息存储过期文件的时间为凌晨4点
deleteWhen=04
# 指定未发生更新的消息存储文件的保留时长为48小时，48小时后过期，将会被删除
fileReservedTime=48
# 指定当前broker为异步复制master
brokerRole=ASYNC_MASTER
# 指定刷盘策略为异步刷盘
flushDiskType=ASYNC_FLUSH
# 指定Name Server的地址
namesrvAddr=192.168.59.164:9876;192.168.59.165:9876
```

```properties
###	broker-b-s.properties
brokerClusterName=DefaultCluster
# 指定这是另外一个master-slave集群
brokerName=broker-b
# slave的brokerId为非0
brokerId=1
deleteWhen=04
fileReservedTime=48
# 指定当前broker为slave
brokerRole=SLAVE
flushDiskType=ASYNC_FLUSH
namesrvAddr=192.168.59.164:9876;192.168.59.165:9876
# 指定Broker对外提供服务的端口，即Broker与producer与consumer通信的端口。默认10911。由于当前主机同时充当着master1与slave2，而前面的master1使用的是默认端口。这里需要将这两个端口加以区分，以区分出master1与slave2
listenPort=11911
# 指定消息存储相关的路径。默认路径为~/store目录。由于当前主机同时充当着master1与slave2，master1使用的是默认路径，这里就需要再指定一个不同路径
storePathRootDir=~/store-s
storePathCommitLog=~/store-s/commitlog
storePathConsumeQueue=~/store-s/consumequeue
storePathIndex=~/store-s/index
storeCheckpoint=~/store-s/checkpoint
abortFile=~/store-s/abort
```

## 8.mqadmin命令

大概就是一些可以通过控制台执行的命令

# 三、RocketMQ工作原理

## 1.消息的生产

### **消息的生产过程**

Producer可以将消息写入到某Broker中的某Queue中，其经历了如下过程：

- Producer发送消息之前，会先向NameServer发出**获取消息Topic的路由信息**的请求
- NameServer返回该**Topic的路由表**及**Broker列表**
- Producer根据代码中指定的Queue选择策略，从Queue列表中选出一个队列，用于后续存储消息
- Produer对消息做一些特殊处理，例如，消息本身超过4M，则会对其进行压缩
- Producer向选择出的Queue所在的Broker发出RPC请求，将消息发送到选择出的Queue

$\textcolor{RoyalBlue}{这里主要搞清楚路由表和Broker列表的结构}$

> **路由表**：实际是一个**Map**，**key**为Topic名称，**value**是一个QueueData实例列表。
>
> QueueData并不是一个Queue对应一个QueueData，而是一个Broker中该Topic的所有Queue对应一个QueueData。
>
> 即，只要涉及到该Topic的Broker，一个Broker对应一个QueueData。QueueData中包含brokerName。简单来说，路由表的key为Topic名称，value则为所有涉及该Topic的BrokerName列表。$\textcolor{RoyalBlue}{(这里就是一个总结)}$
>
> 
>
> **Broker列表**：其实际也是一个**Map**。**key**为brokerName，**value**为BrokerData。
>
> 一个Broker对应一个BrokerData实例，对吗？
>
> 不对。**一套brokerName名称相同的Master-Slave小集群对应一个BrokerData**。BrokerData中包含brokerName及一个map。该map的key为brokerId，value为该broker对应的地址。brokerId为0表示该broker为Master，非0表示Slave。
>
> <img src="RocketMQ简单笔记.assets\image-20220406154601879.png" alt="image-20220406154601879" style="zoom: 50%;" />
>
> $\textcolor{RoyalBlue}{(Broker的示意图)}$

### Queue选择算法

对于无序消息，其Queue选择算法，也称为消息投递算法，常见的有两种：

**轮询算法**

默认选择算法。该算法保证了每个Queue中可以均匀的获取到消息。

> 该算法存在一个问题：由于某些原因，在某些Broker上的Queue可能投递延迟较严重。从而导致Producer的缓存队列中出现较大的消息积压，影响消息的投递性能。$\textcolor{RoyalBlue}{(因为有的Queue延迟较大，半天才能写进去，所以生产者要一直等到确认写入)}$

**最小投递延迟算法**

该算法会统计每次消息投递的时间延迟，然后根据统计出的结果将消息投递到时间延迟最小的Queue。

如果延迟相同，则采用轮询算法投递。该算法可以有效提升消息的投递性能。

> 该算法也存在一个问题：消息在Queue上的分配不均匀。投递延迟小的Queue其可能会存在大量的消息。而对该Queue的消费者压力会增大，降低消息的消费能力，可能会导致MQ中消息的堆积。$\textcolor{RoyalBlue}{(因为生产者可能盯着一个Queue一直塞，虽然生产快了，但是到消费者消费的时候可能只有一个消费者在消费)}$

## 2.消息的存储

RocketMQ中的消息存储在本地文件系统中，这些相关文件默认在当前用户主目录下的store目录中。

$\textcolor{RoyalBlue}{(这应该是一个Broker的目录，一个!!!)}$

<img src="RocketMQ简单笔记.assets\image-20220406214626834.png" alt="image-20220406214626834" style="zoom:50%;" />

>`abort`：该文件在Broker启动后会自动创建，正常关闭Broker，该文件会自动消失。若在没有启动Broker的情况下，发现这个文件是存在的，则说明之前Broker的关闭是非正常关闭。
>`checkpoint`：其中存储着commitlog、consumequeue、index文件的最后刷盘时间戳
>$\textcolor{Red}{commitlog}$：其中存放着commitlog文件，而**消息**是写在commitlog文件中的
>`config`：存放着Broker运行期间的一些配置数据
>$\textcolor{Red}{consumequeue}$：其中存放着consumequeue文件，**队列**就存放在这个目录中
>`index`：其中存放着消息索引文件indexFile
>`lock`：运行期间使用到的全局资源锁

### **commitlog文件**

又称为mappedFile

#### 目录与文件

commitlog目录中存放着很多的mappedFile文件，当前Broker中的所有消息都是落盘到这些mappedFile文件中的。mappedFile文件大小为1G（小于等于1G），文件名由**20位十进制数**构成，表示当前文件的第一条消息的起始位移偏移量。

> 第一个文件名一定是20位0构成的。因为第一个文件的第一条消息的偏移量commitlog offset为0
> 当第一个文件放满时，则会自动生成第二个文件继续存放消息。
>
> 假设第一个文件大小是1073741820字节（1G = 1073741824字节），则第二个文件名就是00000000001073741820。以此类推，第n个文件名应该是前n-1个文件大小之和。
>
> 一个Broker中所有mappedFile文件的commitlog offset是连续的

需要注意的是，一个Broker中仅包含一个commitlog目录，所有的mappedFile文件都是存放在该目录中的。即无论当前Broker中存放着多少Topic的消息，这些消息都是被顺序写入到了mappedFile文件中的。也就是说，**这些消息在Broker中存放时并没有被按照Topic进行分类存放。**

> mappedFile文件是顺序读写的文件，所有其访问效率很高

#### 消息单元

![image-20220407002819965](RocketMQ简单笔记.assets\image-20220407002819965.png)

mappedFile文件内容由一个个的消息单元构成。每个消息单元中包含消息总长度MsgLen、消息的物理位置physicalOffset、消息体内容Body、消息体长度BodyLength、消息主题Topic、Topic长度TopicLength、消息生产者BornHost、消息发送时间戳BornTimestamp、消息所在的队列QueueId、消息在Queue中存储的偏移量QueueOffset等近20余项消息相关属性。

> 一个mappedFile文件中第m+1个消息单元的commitlog offset偏移量
>
> L(m+1) = L(m) + MsgLen(m) (m >= 0)

### consumequeue

<img src="RocketMQ简单笔记.assets\image-20220407002950097.png" alt="image-20220407002950097" style="zoom:50%;" />

#### 目录与文件

<img src="RocketMQ简单笔记.assets\image-20220407003119399.png" alt="image-20220407003119399" style="zoom:67%;" />

为了提高效率，会为每个Topic在~/store/consumequeue中创建一个目录，目录名为Topic名称。在该Topic目录下，会再为每个该Topic的Queue建立一个目录，目录名为queueId。每个目录中存放着若干consumequeue文件，consumequeue文件是commitlog的索引文件，可以根据consumequeue定位到具体的消息。$\textcolor{RoyalBlue}{(大概的层级结构：topic/queue/file)}$

> consumequeue文件名也由20位数字构成，表示当前文件的第一个索引条目的起始位移偏移量。与mappedFile文件名不同的是，其后续文件名是固定的。因为consumequeue文件大小是固定不变的。

#### 索引条目

<img src="RocketMQ简单笔记.assets\image-20220407003256413.png" alt="image-20220407003256413" style="zoom: 67%;" />

每个consumequeue文件可以包含**30w**个索引条目，每个索引条目包含了三个消息重要属性：消息在mappedFile文件中的**偏移量CommitLog Offset**、**消息长度**、**消息Tag的hashcode值**。这三个属性占**20个字节**，所以每个文件的大小是固定的**30w * 20字节。**

> 一个consumequeue文件中所有消息的Topic一定是相同的。但每条消息的Tag可能是不同的。

### 对文件的读写

<img src="RocketMQ简单笔记.assets\image-20220407003515950.png" alt="image-20220407003515950" style="zoom:67%;" />

#### 消息写入

一条消息进入到Broker后经历了以下几个过程才最终被持久化。

1. Broker根据queueId，获取到该消息对应索引条目要在consumequeue目录中的写入偏移量，即QueueOffset

2. 将queueId、queueOffset等数据，与消息一起封装为消息单元

3. 将消息单元写入到commitlog

   $\textcolor{RoyalBlue}{(这里到底是先写到consumequeue还是commitlog没有搞得太明白，个人认为是commitlog)}$

   $\textcolor{RoyalBlue}{(不然我目前把consumequeue理解成一个索引，然后先写索引后写数据感觉很奇怪，万一写的时候G了，索引读出来岂不是一个垃圾数据？)}$

4. 同时，形成消息索引条目

5. 将消息索引条目分发到相应的consumequeue

#### 消息拉取

 当Consumer来拉取消息时会经历以下几个步骤

1. Consumer获取到其要消费消息所在Queue的**消费偏移量offset** ，计算出其要消费消息的消息offset

   > 消费offset即消费进度，consumer对某个Queue的消费offset，即消费到了该Queue的第几条消息
   > 消息offset = 消费offset + 1

2. Consumer向Broker发送拉取请求，其中会包含其要拉取消息的Queue、消息offset及消息Tag。

3. Broker计算在该consumequeue中的queueOffset。

   > queueOffset = 消息offset * 20字节

4. 从该queueOffset处开始向后查找第一个指定Tag的索引条目。

5. 解析该索引条目的前8个字节，即可定位到该消息在commitlog中的commitlog offset

6. 从对应commitlog offset中读取消息单元，并发送给Consumer

#### 性能提升