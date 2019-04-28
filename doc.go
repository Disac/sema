/**
The semaphore in plan9
1. Introduction
Semaphores are now more than 40 years old. Edsger W. Dijkstra described them in EWD 74 [Dijkstra, 1965 (in Dutch)].
A semaphore is a non-negative integer with two operations on it, P and V. The origin of the names P and V is unclear.
In EWD 74, Dijkstra calls semaphores seinpalen (Dutch for signalling posts) and associates V with verhoog (increment/increase) and P with prolaag,
a non-word resembling verlaag(decrement/decrease).
He continues, ‘‘Opm. 2. Vele seinpalen nemen slechts de waarden 0 en 1 aan. In dat geval fungeert de V−operatie als ‘baanvak vrijgeven’;
de P−operatie, de tentatieve passering, kan slechts voltooid worden, als de betrokken seinpaal(of seinpalen) op veilig staat en passering impliceert dan een op onveilig zetten.’’
(Remark 2. Many signals assume only the values 0 and 1. In that case the V-operation functions as release block; the P-operation, the tentative passing,
can only be completed, if the signal (or signals) involved indicates clear, and passing then implies setting it to stop.)
Thus, it may be that P and V were inspired by the railway terms passeer (pass) and verlaat (leave).
We discard the railway terminology and use the language of locks: P is semacquire and V is semrelease. The C declarations are:
	int semacquire(long *addr, int block);
	long semrelease(long *addr, long count);
Semacquire waits for the semaphore value *addr to become positive and then decrements it, returning 1;
if the block flag is zero, semacquire returns 0 rather than wait. If semacquire is interrupted, it returns 1. Semrelease increments the semaphore value by the specified count.
Plan 9 [Pike et al., 1995] has traditionally used a different synchronization mechanism, called rendezvous. Rendezvous is a symmetric mechanism;
that is, it does not assign different roles to the two processes involved. The first process to call rendezvous will block until the second does.
In contrast, semaphores are an asymmetric mechanism: the process executing semacquire can block but the process executing semrelease is guaranteed not to.
We added semaphores to Plan 9 to provide a way for a real-time process to wake up another process without running the risk of blocking.
Since then, we have also used semaphores for efficient process wakeup and locking.
1. 介绍
	信号量拥有超过40年的历史。Edsger W. Dijkstra描述了他们在EWD 74 [Dijkstra, 1965 (in Dutch)]。信号量是一个非负整数，并且拥有2个操作，P和V。P和V的名字起源不清楚。
在EWD 74中，Dijkstra称信号量是"seinpalen"（荷兰语：信号发布）并将V称"verhoog（递增/增加）"和P称"prolaag"（一个非类似于"verlaag（递减/减小）"）。
他称，”Opm. 2. Vele seinpalen nemen slechts de waarden 0 en 1 aan. In dat geval fungeert de V−operatie als ‘baanvak vrijgeven’;
de P−operatie, de tentatieve passering, kan slechts voltooid worden, als de betrokken seinpaal(of seinpalen) op veilig staat en passering impliceert dan een op onveilig zetten.“
（此段为荷兰语：备注2。 许多信号仅有值为0和1。在这种情况下，V操作作为释放块; P操作，暂定的传递，只能是完成后，如果所涉及的信号（或信号量）指示清楚，则通过则暗示设置它停止。）
因此，P和V可能受到了铁路术语的启发代表passeer（通过）和verlaat（离开）。
	我们丢弃铁路术语并使用锁的语言：P是semacquire和V是semrelease。 C声明是：
	int semacquire(long *addr, int block);
	long semrelease(long *addr, long count);
Semacquire等待信号量值*addr变为正值，然后递减它，返回1; 如果block标志为零，则semacquire返回0而不是等待。如果semacquire被中断，它返回1。
Semrelease增加信号量值*addr，增量：count。
	Plan9[Pike et al。，1995]传统上使用了一种不同的同步机制，称为"rendezvous"（集合点）。Rendezvous是一种对称机制，也就是说，它并没有分配不同的角色给两个当前的进程。
调用rendezvous的第一个进程将阻塞，直到第二个进程。相反，信号量是一种非对称机制：当进程执行semacquire的时候会被阻塞，但是执行semrelease的进程保证不会。我们在计划9中添加了信号量，
为实时进程提供了一种唤醒另一个进程的方法，而不会产生阻塞风险。从那时起，我们还使用信号量进行有效的进程唤醒和锁定。

2. Hardware primitives
The implementations in this paper assume hardware support for atomic read-modifywrite operations on a single memory location. The fundamental operation is compare
and swap, which behaves like this C function cas, but executes atomically:
	int
	cas(long *addr, long old, long new)
	{
		// Executes atomically.
		if(*addr != old)
			return 0;
		*addr = new;
		return 1;
	}
In one atomic operation, cas checks whether the value *addr is equal to old and, if so, changes it to new. It returns a flag telling whether it changed *addr.
Of course, cas is not implemented in C. Instead, we must implement it using special hardware instructions. All modern processors provide a way to implement compare and swap.
The x86 architecture (since the 486) provides a direct compare and swap instruction, CMPXCHG.
Other processorsincluding the Alpha, ARM, MIPS, and PowerPCprovide a pair of instructions called load linked (LL) and store conditional (SC).
The LL instruction reads from a memory location, and SC writes to a memory location only if (1) it was the memory location used in the last LL instruction,
and (2) that location has not been changed since the LL. On those systems, compare and swap can be implemented in terms of LL and SC.
The implementations also use an atomic addition operation xadd that atomically adds to a value in memory, returning the new value. We dont need additional hardware support for xadd,
since it can be implemented using cas:
	long
	xadd(long *addr, long delta)
	{
		long v;
		for(;;){
			v = *addr;
			if(cas(addr, v, v+delta))
				return v+delta;
		}
	}
2. 硬件原函数
	本文中的实现假设硬件支持单个内存位置上的原子读取 - 修改写操作。基本操作是比较并交换，其行为类似于此C函数cas，但以原子方式执行：
	cas(long *addr, long old, long new)
	{
		// 执行源自操作
		if(*addr != old)
			return 0;
		*addr = new;
		return 1;
	}
在一个原子操作中，cas检查值*addr是否等于old，如果是，将其更改为new。它返回一个标志，告诉它是否更改了*addr。
	当然，cas不是用C实现的。我们必须使用特殊的硬件指令来实现它。所有现代处理器都提供了实现比较并交换的方法。
x86架构（自486开始）提供直接比较和交换指令CMPXCHG。其他处理器包括Alpha，ARM，MIPS和PowerPC提供一对称为"加载链接（LL）"和"存储条件（SC）"的指令。
LL指令从存储器位置读取，SC写入存储器位置仅当（1）它是最后一条LL指令中使用的存储单元，以及（2）该位置时LL以来没有改变过。
在这些系统上，可以根据LL和SC实现比较和交换。这些实现还使用原子级的原子加法运算xadd在内存中添加一个值，返回新值。
我们不需要额外的硬件支持xadd，因为它可以使用cas实现：
	long
	xadd(long *addr, long delta)
	{
		long v;
		for(;;){
			v = *addr;
			if(cas(addr, v, v+delta))
				return v+delta;
		}
	}

3. User−space semaphores
We implemented semacquire and semrelease as kernel-provided system calls. For efficiency, it is useful to have a semaphore implementation that,
if there is no contention, can run entirely in user space, only falling back on the kernel to handle contention.
Figure 1 gives the implementation. The user space semaphore, a Usem, consists of a user-level semaphore value u and a kernel value k:
typedef struct Usem Usem;
struct Usem {
	long u;
	long k;
};
When u is non-negative, it represents the actual semaphore value. When u is negative,
the semaphore has value zero: acquirers must wait on the kernel semaphore k and releasers must wake them up.
	void
	usemacquire(Usem *s)
	{
		if(xadd(&s−>u, −1) < 0)
			while(semacquire(&s−>k, 1) < 0){
				// 中断，重试
			}
	}
	void
	usemrelease(Usem *s)
	{
		if(xadd(&s−>u, 1) <= 0)
			semrelease(&s−>k, 1);
	}
If the semaphore is uncontended, the xadd in usemacquire will return a non-negative value, avoiding the kernel call.
Similarly, the xadd in usemrelease will return a positive value, also avoiding the kernel call.
3. 用户空间信号量
	我们实现了semacquire和semrelease作为内核提供的系统调用。
为了提高效率，有一个信号量实现很有用，如果没有争用，它可以完全在用户空间中运行，只能回退到内核来处理争用。
图1给出了实现。 用户空间信号量，即Usem，由用户级信号量值u和内核值k组成：
	typedef struct Usem Usem;
	struct Usem {
		long u;
		long k;
	};
当u为非负数时，它表示实际的信号量值。当u为负数时，信号量的值为零：获取者必须等待内核信号量k并且释放者必须将它们唤醒。
	void
	usemacquire(Usem *s)
	{
		if(xadd(&s−>u, −1) < 0)
			while(semacquire(&s−>k, 1) < 0){
				// 中断，重试
			}
	}
	void
	usemrelease(Usem *s)
	{
		if(xadd(&s−>u, 1) <= 0)
			semrelease(&s−>k, 1);
	}
如果信号量是无竞争的，则usemacquire中的xadd将返回非负值，从而避免内核调用。同样，usemrelease中的xadd将返回一个正值，同时也避免了内核调用。

4. Thread Scheduling
In the Plan 9 thread library, a program is made up of a collection of processes sharing memory. A thread is a coroutine assigned to a particular process. Within a process, threads schedule cooperatively.
Each process manages the threads assigned to it, and the process schedulers run almost independently.
The one exception is that a thread in one process might go to sleep (for example, waiting on a channel operation) and be woken up by a thread in a different process.
The two processes need a way to coordinate, so that if the first has no runnable threads, it can go to sleep in the kernel, and then the second process can wake it up.
The standard Plan 9 thread library uses rendezvous to coordinate between processes.
The processes share access to each others scheduling queues: one process is manipulating anothers run queue.
The processes must also share a flag protected by a spin lock to coordinate, so that either both processes decide to call rendezvous or neither does.
For the real-time thread library, we wanted to remove as many sources of blocking as possible, including these locks.
We replaced the locked run queue with a nonblocking array-based implementation of a producer/consumer queue.
That implementation is beyond the scope of this paper. After making that change, the only lock remaining in the scheduler was the one protecting the "whether to rendezvous" flag.
To eliminate that one, we replaced the rendezvous with a user-space semaphore counting the number of threads on the queue.
To wait for a thread to run, the processs scheduler decrements the semaphore. If the run queue is empty, the usemacquire will block until it is not. Having done so,
it is guaranteed that there is a thread on the run queue:
	// Get next thread to run
	static Thread*
	runthread(void)
	{
		Proc *p;
		p = thisproc();
		usemacquire(&p−>nready);
		return qget(&p−>ready);
	}
Similarly, to wake up a thread (even one in another process), it suffices to add the thread to its processs run queue and then increment the semaphore:
	// Wake up thread t to run in its process.
	static void
	wakeup(Thread *t)
	{
		Proc *p;
		p = t−>p;
		qput(&p−>ready, t);
		usemrelease(&p−>nready);
	}
This implementation removes the need for the flag and the lock; more importantly, the process executing threadwakeup is guaranteed never to block, because it executes usemrelease, not usemacquire.
4. 线程调度
	在Plan 9线程库中，程序由共享内存的进程集合组成。线程是分配给特定进程的协程。在一个进程中，线程协同合作。每个进程管理分配给它的线程，
进程调度程序几乎独立运行。一个例外是一个进程中的线程可能会进入休眠状态（例如，等待通道操作）并被另一个进程中的线程唤醒。这两个进程需要一种协调方式，
因此如果第一个进程没有可运行的线程，它可以在内核中进入休眠状态，然后第二个进程可以将其唤醒。
	标准Plan9线程库使用rendezvous（集合）来协调进程之间的协调。进程共享对彼此调度队列的访问：一个进程正在操纵另一个运行队列。进程还必须共享一个由自旋锁保护的标志来进行协调，
这样两个进程都决定调用rendezvous（集合），或者两者都没有。
	对于实时线程库，我们希望尽可能多地删除阻塞源，包括这些锁。我们用生产者/消费者队列的非阻塞基于数组的实现替换了已锁定的运行队列。该实现超出了本文的范围。进行更改后，
调度程序中剩余的唯一锁定是保护“是否rendezvous（集合）”标志的锁。为了消除这一点，我们用一个计算队列中线程数的用户空间信号量替换了rendezvous。
	要等待线程运行，进程的调度程序会递减信号量。如果运行队列为空，则usemacquire将阻塞，直到它不为空。 完成后，可以保证运行队列中有一个线程：
	// 获取下一个要运行的线程
	static Thread*
	runthread(void)
	{
		Proc *p;
		p = thisproc();
		usemacquire(&p−>nready);
		return qget(&p−>ready);
	}
同样，要唤醒一个线程（即使在另一个进程中），只需将线程添加到其进程的运行队列，然后递增信号量即可：
	// 唤醒线程t以在其进程中运行。
	static void
	wakeup(Thread *t)
	{
		Proc *p;
		p = t−>p;
		qput(&p−>ready, t);
		usemrelease(&p−>nready);
	}
这种实现消除了对标志和锁的需要; 更重要的是，执行threadwakeup的进程保证永远不会阻塞，因为它执行usemrelease，而不是usemacquire。

5. Replacing spin locks
The Plan 9 user-level Lock implementation is an adapted version of the one used in the kernel. A lock is represented by an integer value: 0 is unlocked, non-zero is locked.
A process tries to grab the lock by using a test-and-set instruction to check whether the value is 0 and, if so, set it to a non-zero value. If the lock is unavailable, the process loops,
trying repeatedly. In a multiprocessor kernel, this is a fine lock implementation: the lock is held by another processor, which will unlock it soon.
In user space, this implementation has bad interactions with the scheduler: if the lock is held by another process that has been preempted, spinning for the lock will not accomplish anything.
The user-level lock implementation addresses this by rescheduling itself (with sleep(0)) between attempts after the first thousand unsuccessful attempts.
Eventually it backs off more, sleeping for milliseconds at a time between lock attempts.
We replaced these spin locks with a semaphore-based implementation. Using semaphores allows the process to tell the kernel exactly what it is waiting for,
avoiding bad interactions with the scheduler like the one above. The semaphore-based implementation represents a lock as two values, a user-level key and a kernel semaphore:
	struct Lock
	{
		long key;
		long sem;
	};
The key counts the number of processes interested in holding the lock, including the one that does hold it. Thus if key is 0, the lock is unlocked. If key is 1, the lock is held.
If key is larger than 1, the lock is held by one process and there are key1 processes waiting to acquire it. Those processes wait on the semaphore sem.
	void
	lock(Lock *l)
	{
		if(xadd(&l−>key, 1) == 1)
			return; // changed from 0 −> 1: we hold lock
		// otherwise wait in kernel
		while(semacquire(&l−>sem, 1) < 0){
			// interrupted; try again
		}
	}
	void
	unlock(Lock *l)
	{
		if(xadd(&l−>key, −1) == 0)
			return; // changed from 1 −> 0: no contention
		semrelease(&l−>sem, 1);
	}
Like the user-level semaphore implementation described above, the lock implementation handles the uncontended case without needing to enter the kernel.
The one significant difference between the user-level semaphores above and the semaphore-based locks described here is the interpretation of the user-space value.
Plan 9 convention requires that a zeroed Lock structure be an unlocked lock. In contrast, a zeroed Usem structure is analogous to a locked lock: a usemacquire on a zeroed Usem will block.
5. 更换旋锁
	Plan 9用户级Lock实现是内核中使用的版本的改编版本。锁由整数值表示：0解锁，非零锁定。进程尝试使用test-and-set指令来获取锁，以检查值是否为0，如果是这样，请将其设置为非零值。如果锁不可用，则进程循环，反复尝试。
在多处理器内核中，这是一个精确的锁实现：锁由另一个处理器保持，它将很快解锁。
在用户空间中，此实现与调度程序的交互不好：如果锁被另一个被抢占的进程持有，那么为锁进行旋转将无法完成任何操作。
用户级锁实现通过在第一千次尝试失败后的尝试之间重新安排自身（使用sleep（0））来解决此问题。最终它会更多地退回，在锁定尝试之间一次睡眠1毫秒。
	我们用基于信号量的实现替换了这些自旋锁。使用信号量允许进程准确地告诉内核它正在等待什么，避免与上述调度程序的错误交互。基于信号量的实现的锁有两个值，即用户级key和内核信号量：
	struct Lock
	{
		long key;
		long sem;
	};
key是有兴趣持有锁的进程数，包括那个持有它的人。因此，如果key为0，则锁定被解锁。 如果key为1，则保持锁定。如果key大于1，则锁由一个进程保持，并且有关键1进程等待获取它。 这些进程等待信号量sem。
	void
	lock(Lock *l)
	{
		if(xadd(&l−>key, 1) == 1)
			return; // changed from 0 −> 1: we hold lock
		// otherwise wait in kernel
		while(semacquire(&l−>sem, 1) < 0){
			// interrupted; try again
		}
	}
	void
	unlock(Lock *l)
	{
		if(xadd(&l−>key, −1) == 0)
			return; // changed from 1 −> 0: no contention
		semrelease(&l−>sem, 1);
	}
与上面描述的用户级信号量实现一样，锁实现处理无争用的情况而无需进入内核。上面的用户级信号量和此处描述的基于信号量的锁之间的一个显着差异是对用户空间值的解释。
Plan 9约定要求将零值锁结构作为未锁定。相比之下，归零的Usem结构类似于已锁定的锁：对归零的Usem的usemacquire将阻塞。

6. Kernel Implementation of Semaphores
Inside the Plan 9 kernel, there are two kinds of locks: the spin lock Lock spins until the lock is available, and the queuing lock QLock reschedules the current process until the lock is available.
Because accessing user memory might cause a lengthy page fault, the kernel does not allow a process to hold a Lock while accessing user memory.
Since the semaphore is stored in user memory, then, the obvious implementation is to acquire a QLock, perform the semaphore operations, and then release it.
Unfortunately, this implementation could cause semrelease to reschedule while acquiring the QLock, negating the main benefit of semaphores for real-time processes.
A more complex implementation is needed. This section documents the implementation. It is not necessary to understand the rest of the paper and can be skipped on first reading.
Each semacquire call records its parameters in a Sema data structure and adds it to a list of active calls associated with a particular Segment (a shared memory region).
The Sema structure contains a kernel Rendez for use by sleep and wakeup (see [Pike et al.,1991]), the address, and a waiting flag:
	struct Sema
	{
		Rendez;
		long *addr;
		int waiting;
		Sema *next;
		Sema *prev;
	};
The list is protected by a Lock, which cannot cause the process to reschedule. The
semaphore value *addr is stored in user memory. Thus, we can access the list only
when holding the lock and we can access the semaphore value only when not holding
the lock. The helper functions
	void semqueue(Segment *s, long *addr, Sema *p);
	void semdequeue(Segment *s, long *addr, Sema *p);
	void semwakeup(Segment *s, long *addr, int n);
all manipulate the segments list of Sema structures. They acquire the associated Lock,
perform their operations, and release the lock before returning. Semqueue and
semdequeue add p to or remove p from the list. Semwakeup walks the list looking for n
Sema structures with p.waiting set. It clears p.waiting and then wakes up the corresponding process.
Using those helper functions, the basic implementation of semacquire and semrelease is:
	int
	semacquire(Segment *s, long *addr)
	{
		Sema phore;
		semqueue(s, addr, &phore);
		for(;;){
			phore.waiting = 1;
			if(canacquire(addr))
				break;
			sleep(&phore, semawoke);
		}
		semdequeue(s, &phore);
		semwakeup(s, addr, 1);
		return 1;
	}
	long
	semrelease(Segment *s, long *addr, long n)
	{
		long v;
		v = xadd(addr, n);
		semwakeup(s, addr, n);
		return v;
	}
(This version omits the details associated with returning 1 when interrupted and also
with non-blocking calls.)
Semacquire adds a Sema to the segments list and sets phore.waiting. Then it attempts to acquire the semaphore. If it is unsuccessful, it goes to sleep.
To avoid missed wakeups, sleep calls semawoke before committing to sleeping; semawoke simply checks phore.waiting. Eventually, canacquire returns true, breaking out of the loop.
Then semacquire removes its Sema from the list and returns.
The call to semwakeup at the end of semacquire corrects a subtle race that we found using Spin. Suppose process A calls semacquire and the semaphore has value 1.
Semacquire queues its Sema and sets phore.waiting, canacquire succeeds (the semaphore value is now 0), and semacquire breaks out of the loop.
Then process B calls semacquire: it adds itself to the list, fails to acquire the semaphore (the value is 0), and goes to sleep.
Now process C calls semrelease: it increments the semaphore (the value is now 1) and looks for a single Sema in the list to wake up.
It finds As, checks that phore.waiting is set, and then calls the kernel wakeup to wake A. Unfortunately, A never went to sleep.
The wakeup is lost on A, which had already acquired the semaphore. If A simply removed its Sema from the list and returned, the semaphore value would be 1 with B still asleep.
To account for the possibly lost wakeup, A must trigger one extra semwakeup as it returns. This avoids the race, at the cost of an unnecessary (but harmless) wakeup when the race has not happened.
6. 信号量的内核实现
	在Plan 9内核中，有两种锁：自旋锁Lock旋转直到锁可用，排队锁QLock重新安排当前进程直到锁可用。
由于访问用户内存可能会导致冗长的页面错误，因此内核不允许进程在访问用户内存时保持锁定。
由于信号量存储在用户存储器中，因此，明显的实现是获取QLock，执行信号量操作，然后释放它。
不幸的是，这种实现可能导致semrelease在获取QLock时重新安排，从而否定了信号量对实时进程的主要好处。
需要更复杂的实现。本节记录了实现。没有必要理解本文的其余部分，可以在第一次阅读时跳过。（?_?）
	每个semacquire调用将其参数记录在Sema数据结构中，并将其添加到与特定Segment（共享内存区域）相关联的活动调用列表中。
Sema结构包含一个内核Rendez供睡眠和唤醒使用（参见[Pike et al。，1991]），地址和等待标志：
	struct Sema
	{
		Rendez;
		long *addr;
		int waiting;
		Sema *next;
		Sema *prev;
	};
该列表受Lock保护，不会导致进程重新计划。信号量值*addr存储在用户存储器中。因此，我们只有在持有锁时才能访问列表，只有在没有持有锁的情况下我们才能访问信号量值。
辅助函数：
	void semqueue(Segment *s, long *addr, Sema *p);
	void semdequeue(Segment *s, long *addr, Sema *p);
	void semwakeup(Segment *s, long *addr, int n);
所有人操纵Sema结构的Segment。他们获取相关的锁，执行操作，并在返回之前释放锁。
Semqueue和semdequeue在列表中添加p或从中删除p。
Semwakeup走在列表中寻找具有p.waiting set的n个Sema结构。它清除p.waiting然后唤醒相应的进程。
	使用这些辅助函数，semacquire和semrelease的基本实现是：
	int
	semacquire(Segment *s, long *addr)
	{
		Sema phore;
		semqueue(s, addr, &phore);
		for(;;){
			phore.waiting = 1;
			if(canacquire(addr))
				break;
			sleep(&phore, semawoke);
		}
		semdequeue(s, &phore);
		semwakeup(s, addr, 1);
		return 1;
	}
	long
	semrelease(Segment *s, long *addr, long n)
	{
		long v;
		v = xadd(addr, n);
		semwakeup(s, addr, n);
		return v;
	}
（此版本省略了与中断时以及非阻塞呼叫返回1相关的详细信息。）
	Semacquire在Segment中添加了一个Sema并设置了phore.waiting。 然后它试图获取信号量。 如果不成功，它会进入睡眠状态。
为了避免错过唤醒，sleep前会通知semawoke；semawoke只是检查phore.waiting。最终，canacquire返回true，跳出循环。然后semacquire从列表中删除其Sema并返回。
	在调用semacquire结束时对semwakeup的调用纠正了我们使用自选的一个微妙的竞争。假设进程A调用semacquire并且信号量的值为1。Semacquire将其Sema排队并设置phore.waiting，canacquire成功（信号量值现在为0），semacquire跳出循环。
然后进程B调用semacquire：它将自己添加到列表中，无法获取信号量（值为0），然后进入休眠状态。现在进程C调用semrelease：它递增信号量（值现在为1）并在列表中查找单个Sema以唤醒。它找到A，检查phore.waiting是否已设置，
然后调用内核唤醒来唤醒A.不幸的是，A从未进入睡眠状态。在已经获得信号量的A上失去了唤醒。 如果A只是从列表中删除了它的Sema并返回，那么信号量值将为1，而B仍然处于睡眠状态。为了解释可能丢失的唤醒，A必须在返回时触发一次额外的semwakeup。
这样可以避免竞争，但是当竞争没有发生时会以不必要的（但无害的）唤醒为代价。

7. Performance
To measure the cost of semaphore synchronization, we wrote a program in which two processes ping-pong between two semaphores:
	Process 1 blocks on the acquisition of Semaphore 1,
	Process 2 releases Semaphore 1 and blocks on Semaphore 2,
	Process 1 releases Semaphore 2 and blocks on Semaphore 1,
This loop executes a million times. We also timed a program that does two million acquires and two million releases on a semaphore initialized to two million, so that none of the calls would block.
In both cases, there were a total of four million system calls; the ping-pong case adds two million context switches. Table 1 gives the results.
time per system call (microseconds)
processor cpus ping−pong semacquire semrelease
PentiumIII/Xeon, 598 MHz 1 2.18 1.35 1.91
PentiumIII/Xeon, 797 MHz 2 0.887 0.949 1.38
PentiumIV/Xeon, 2196 MHz 4 0.970 1.38 1.84
AMD64, 2201 MHz 2 1.08 0.266 0.326
Table 1 Semaphore system call performance.

time per lock operation (microseconds)
processor cpus spin locks semaphore locks
PentiumIII/Xeon, 598 MHz 1 5.4 5.4
PentiumIII/Xeon, 797 MHz 2 18.2 5.6
AMD64, 2201 MHz 2 22.6 2.5
PentiumIV/Xeon, 2196 MHz 4 43.8 4.9
Table 2 Performance of spin locks versus semaphore locks

Next, we looked at lock performance, comparing the conventional Plan 9 locks from libc to the new ones using semaphores for sleep and wakeup.
We ran Doug McIlroys power series program [McIlroy, 1990], which spends almost all its time in channel communication.
The Plan 9 thread librarys channel implementation uses a single global lock to coordinate all channel activity, inducing a large amount of lock contention.
The application creates a thousand processes and makes 207,631 lock calls. The number of locks (in the semaphore version) that require waiting (i.e., a semacquire is done) varies wildly.
In 20 runs, the smallest number we saw was 127, the largest was 490, and the average was 288.
Table 2 shows the performance results. Surprisingly, the performance difference was most pronounced on multiprocessors.
Naively, one would expect that spinning would have some benefit on multiprocessors whereas it could have no benefit on uniprocessors,
but it turns out that spinning without rescheduling (the first 1000 tries) has no effect on performance. Contention only occurs some 500 or so times,
and the time it takes to spin 500,000 times is in the noise.
The difference between uniprocessors and multiprocessors here is that on uniprocessors, the first sleep(0) will put the process waiting for the lock at the back of the ready queue so that,
by the time it is scheduled again, the lock will likely be available. On multiprocesssors, contention from other processes running simultaneously makes yielding less effective.
It is also likely that the repeated atomic read-modify-write instructions, as in the tight loop of the spin lock, can slow the entire multiprocessor.
The performance of the semaphore-based lock implementation is sometimes much better, and never noticeably worse, than the spin locks.
We will replace the spin lock implementation in the Plan 9 distribution soon.
7. 表现
	为了衡量信号量同步的成本，我们编写了一个程序，其中两个进程在两个信号量之间进行ping-pong：
	Process 1 blocks on the acquisition of Semaphore 1,
	Process 2 releases Semaphore 1 and blocks on Semaphore 2,
	Process 1 releases Semaphore 2 and blocks on Semaphore 1,
这个循环执行了一百万次。 我们还计划了一个程序，该程序在一个初始化为200万的信号量上完成了200万次获取和200万次发布，因此没有任何调用会阻塞。
在这两种情况下，总共有四百万次系统调用; ping-pong案例增加了200万个上下文切换。 表1给出了结果。
每次系统调用的时间（微秒）
processor cpus ping−pong semacquire semrelease
PentiumIII/Xeon, 598 MHz 1 2.18 1.35 1.91
PentiumIII/Xeon, 797 MHz 2 0.887 0.949 1.38
PentiumIV/Xeon, 2196 MHz 4 0.970 1.38 1.84
AMD64, 2201 MHz 2 1.08 0.266 0.326
表1信号量系统调用性能。

每次锁定操作的时间（微秒）
processor cpus spin locks semaphore locks
PentiumIII/Xeon, 598 MHz 1 5.4 5.4
PentiumIII/Xeon, 797 MHz 2 18.2 5.6
AMD64, 2201 MHz 2 22.6 2.5
PentiumIV/Xeon, 2196 MHz 4 43.8 4.9
表2自旋锁与信号量锁的性能。
	接下来，我们研究了锁性能，将libc的传统Plan 9锁与使用信号量进行睡眠和唤醒的新锁进行比较。我们运行了Doug McIlroy的电源系列程序[McIlroy，1990]，它几乎把所有时间花在了频道通信上。
Plan 9线程库的通道实现使用单个全局锁来协调所有通道活动，从而引发大量的锁争用。该应用程序创建了一千个进程并进行了207,631次锁调用。 需要等待的锁（在信号量版本中）的数量（即，完成了一个semacquire）变化很大。
在20次运行中，我们看到的最小数量是127，最大的是490，平均值是288。
	表2显示了性能结果。 令人惊讶的是，性能差异在多处理器上最为明显。天真地，人们会期望旋转对多处理器有一些好处，而它对单处理器没有任何好处，但事实证明，没有重新安排（前1000次尝试）的旋转对性能没有影响。
争用仅发生约500次左右，旋转500,000次所需的时间是噪音。
这里的单处理器和多处理器之间的区别在于，在单处理器上，第一个sleep（0）将使进程等待就绪队列后面的锁定，以便在再次调度时，锁定可能会可用。 在多处理器上，同时运行的其他进程的争用会降低效率。
重复的原子读 - 修改 - 写指令也很可能，如在自旋锁的紧密循环中，可以减慢整个多处理器的速度。基于信号量的锁实现的性能有时比旋转锁更好，并且从未明显更糟。我们将很快取代Plan 9发行版中的自旋锁实现。


8. Comparison with other approaches
Any operating system with cooperating processes must provide an interprocess synchronization mechanism. It is instructive to contrast the semaphores described here with mechanisms in other systems.
Many systemsfor example, BSD, Mach, OS X, and even System V UNIXprovide semaphores [Bach, 1986]. In all those systems, semaphores must be explicitly allocated and deallocated,
making them more cumbersome to use than semacquire and semrelease.
Worse, semaphores in those systems occupy a global id space,
so that it is possible to run the system out of semaphores just by running programs that allocate semaphores but neglect to deallocate them (or crash).
The Plan 9 semaphores identify semaphores by a shared memory location: two processes are talking about the same semaphore if *addr is the same word of physical memory in both.
Further, there is no kernel-resident semaphore state except when semacquire is blocking. This makes the semaphore leaks of System V impossible.
Linux provides a lower-level system call named futex [Franke and Russell, 2002]. Futex is essentially "compare and sleep", making it a good match for compare and swap-based algorithms.
Futex also matches processes based on shared physical memory, avoiding the System V leak problem. Because futex only provides "compare and sleep"and"wakeup",
futex-based algorithms are required to handle the uncontended cases in user space, like our user-level semaphore and new lock implementations do.
This makes futex-based implementations efficient; unfortunately, they are also quite subtle. The original example code distributed with futexes was wrong;
a correct version was only published a year later [Drepper, 2003]. In contrast, semaphores are less general but easier to understand and to use correctly.
8.与其他方法的比较
	具有协作进程的任何操作系统都必须提供进程间同步机制。 将这里描述的信号量与其他系统中的机制进行对比是有益的。许多系统，例如BSD，Mach，OS X，甚至System V UNIX都提供信号量[Bach，1986]。
在所有这些系统中，必须明确分配和释放信号量，使它们比semacquire和semrelease使用起来更麻烦。更糟糕的是，这些系统中的信号量占据了全局身份空间，这样就可以通过运行分配信号量但忽略解除分配信号（或崩溃）的程序来运行信号量系统。
Plan 9信号量通过共享内存位置识别信号量：如果*addr与两者中的物理内存相同，则两个进程正在讨论相同的信号量。此外，除了semacquire阻塞之外，没有内核驻留的信号量状态。 这使得System V的信号量泄漏变得不可能。
	Linux提供了一个名为futex的低级系统调用[Franke and Russell，2002]。 Futex本质上是“比较和休眠”，使其成为比较和基于交换的算法的良好匹配。Futex还基于共享物理内存匹配进程，避免了System V泄漏问题。
因为futex只提供“比较和睡眠”和“唤醒”，需要基于futex的算法来处理用户空间中的无竞争情况，例如我们的用户级信号量和新的锁实现。这使基于futex的实现变得高效; 不幸的是，它们也很微妙。
与futexes一起分发的原始示例代码是错误的; 一个正确的版本仅在一年后发布[Drepper，2003]。 相比之下，信号量不太通用，但更容易理解和正确使用。
*/
package sema
