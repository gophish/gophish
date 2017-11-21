# Add-on Installation Instructions
Enable the plugin by using the extraPlugins configuration setting
```
config.extraPlugins = 'videosnapshot';
```

If you use `sourcedialog` plugin, you need to enable `allowedContent` option
```
config.allowedContent = true;
```


#  Add item with custom toolbar
```
config.toolbar = [
	...
	{ name: 'other', items: ['videosnapshot'] }
]
```
