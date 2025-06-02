So Hon is basically just a Simple BookTracer app...

But there is some new features I'm adding: which is goal chasing system. So basically you can create a goal (how much page you want to target and when the deadline is).

You can make a progress with reading a book, and maybe write a note.

If you finished the goal in time, it will sends you a congratulation message via email. And if you're not, it will sends a condolence message :laughingCatEmote.

If you wonder how the system can sends a message when the goal's deadline is coming, I'm using RabbitMQ's delayed message plugins, it allows me to delay a message in a period of time before it will passed to a queue, the consumer then takes the message inside queue and sends an mail.

For the goal completion message, I also make it using RabbitMQ and the flow is much simpler.


Update Note:
- Dockerized in coming
- Some feat are unstable also