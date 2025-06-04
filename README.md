So Hon is basically just a Simple Book Tracer app.

You can make account, register a book, make progress.

But... there is some features I'm adding: which is goal chasing system. So basically you can create a goal (how much page you want to target and when the deadline is).

You can make a progress with reading a book, and maybe write a note.

If you finished the goal in time, it will sends you a congratulation message via email. And if you're not, it will sends a condolence message :laughingCatEmote.

If you wonder how the system can sends a message when the goal's deadline coming. The answer is RabbitMQ, precisely... RabbitMQ's delayed message plugin, it allows me to hold/delay a message before it sent to a queue. Then, the consumer takes that message inside the queue and processes it.

Note: 
This repo is just my playground to escape RabbitMQ tutorial hell.

Update:
- Dockerized app incoming.
- Docs also incoming