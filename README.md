# Usage
```console
# Add new songs:
$ songs --insert 'Le Wanski - Trauma (N'to - Trauma Remix)'
$ songs --insert 'MALO - March of Progress'
$ # List songs (newest additions to oldest):
$ songs
MALO - March of Progress
Le Wanski - Trauma (N'to - Trauma Remix)
$ # Register, that you are listening to a song (you might use this in
$ # scripts):
$ songs --register 'Le Wanski - Trauma (N'to - Trauma Remix)'
$ # List those songs first, that you listened to the most:
$ songs --favourites
Le Wanski - Trauma (N'to - Trauma Remix)
MALO - March of Progress
$ # List songs by frecency (songs you lately listened to a lot):
$ songs --frecent
Le Wanski - Trauma (N'to - Trauma Remix)
MALO - March of Progress
$ # List songs, that you often listen to before or after listening to
$ # the given song:
$ songs --similar 'Le Wanski - Trauma (N'to - Trauma Remix)'
MALO - March of Progress
```
