/**
 * @license Copyright (c) 2003-2017, CKSource - Frederico Knabben. All rights reserved.
 * For licensing, see LICENSE.md or http://ckeditor.com/license
 */

/**
 * @fileOverview The "Notification Aggregator" plugin.
 *
 */

( function() {
	'use strict';

	CKEDITOR.plugins.add( 'notificationaggregator', {
		requires: 'notification'
	} );

	/**
	 * An aggregator of multiple tasks (progresses) which should be displayed using one
	 * {@link CKEDITOR.plugins.notification notification}.
	 *
	 * Once all the tasks are done, it means that the whole process is finished and the {@link #finished}
	 * event will be fired.
	 *
	 * New tasks can be created after the previous set of tasks is finished. This will continue the process and
	 * fire the {@link #finished} event again when it is done.
	 *
	 * Simple usage example:
	 *
	 *		// Declare one aggregator that will be used for all tasks.
	 *		var aggregator;
	 *
	 *		// ...
	 *
	 *		// Create a new aggregator if the previous one finished all tasks.
	 *		if ( !aggregator || aggregator.isFinished() ) {
	 *			// Create a new notification aggregator instance.
	 *			aggregator = new CKEDITOR.plugins.notificationAggregator( editor, 'Loading process, step {current} of {max}...' );
	 *
	 *			// Change the notification type to 'success' on finish.
	 *			aggregator.on( 'finished', function() {
	 *				aggregator.notification.update( { message: 'Done', type: 'success' } );
	 *			} );
	 *		}
	 *
	 *		// Create 3 tasks.
	 *		var taskA = aggregator.createTask(),
	 *			taskB = aggregator.createTask(),
	 *			taskC = aggregator.createTask();
	 *
	 *		// At this point the notification contains a message: "Loading process, step 0 of 3...".
	 *
	 *		// Let's close the first one immediately.
	 *		taskA.done(); // "Loading process, step 1 of 3...".
	 *
	 *		// One second later the message will be: "Loading process, step 2 of 3...".
	 *		setTimeout( function() {
	 *			taskB.done();
	 *		}, 1000 );
	 *
	 *		// Two seconds after the previous message the last task will be completed, meaning that
	 *		// the notification will be closed.
	 *		setTimeout( function() {
	 *			taskC.done();
	 *		}, 3000 );
	 *
	 * @since 4.5
	 * @class CKEDITOR.plugins.notificationAggregator
	 * @mixins CKEDITOR.event
	 * @constructor Creates a notification aggregator instance.
	 * @param {CKEDITOR.editor} editor
	 * @param {String} message The template for the message to be displayed in the notification. The template can use
	 * the following variables:
	 *
	 * * `{current}` &ndash; The number of completed tasks.
	 * * `{max}` &ndash; The number of tasks.
	 * * `{percentage}` &ndash; The progress (number 0-100).
	 *
	 * @param {String/null} [singularMessage=null] An optional template for the message to be displayed in the notification
	 * when there is only one task remaining.  This template can use the same variables as the `message` template.
	 * If `null`, then the `message` template will be used.
	 */
	function Aggregator( editor, message, singularMessage ) {
		/**
		 * @readonly
		 * @property {CKEDITOR.editor} editor
		 */
		this.editor = editor;

		/**
		 * Notification created by the aggregator.
		 *
		 * The notification object is modified as aggregator tasks are being closed.
		 *
		 * @readonly
		 * @property {CKEDITOR.plugins.notification/null}
		 */
		this.notification = null;

		/**
		 * A template for the notification message.
		 *
		 * The template can use the following variables:
		 *
		 * * `{current}` &ndash; The number of completed tasks.
		 * * `{max}` &ndash; The number of tasks.
		 * * `{percentage}` &ndash; The progress (number 0-100).
		 *
		 * @private
		 * @property {CKEDITOR.template}
		 */
		this._message = new CKEDITOR.template( message );

		/**
		 * A template for the notification message used when only one task is loading.
		 *
		 * Sometimes there might be a need to specify a special message when there
		 * is only one task loading, and to display standard messages in other cases.
		 *
		 * For example, you might want to show a message "Translating a widget" rather than
		 * "Translating widgets (1 of 1)", but still you would want to have a message
		 * "Translating widgets (2 of 3)" if more widgets are being translated at the same
		 * time.
		 *
		 * Template variables are the same as in {@link #_message}.
		 *
		 * @private
		 * @property {CKEDITOR.template/null}
		 */
		this._singularMessage = singularMessage ? new CKEDITOR.template( singularMessage ) : null;

		// Set the _tasks, _totalWeights, _doneWeights, _doneTasks properties.
		this._tasks = [];
		this._totalWeights = 0;
		this._doneWeights = 0;
		this._doneTasks = 0;

		/**
		 * Array of tasks tracked by the aggregator.
		 *
		 * @private
		 * @property {CKEDITOR.plugins.notificationAggregator.task[]} _tasks
		 */

		/**
		 * Stores the sum of declared weights for all contained tasks.
		 *
		 * @private
		 * @property {Number} _totalWeights
		 */

		/**
		 * Stores the sum of done weights for all contained tasks.
		 *
		 * @private
		 * @property {Number} _doneWeights
		 */

		/**
		 * Stores the count of tasks done.
		 *
		 * @private
		 * @property {Number} _doneTasks
		 */
	}

	Aggregator.prototype = {
		/**
		 * Creates a new task that can be updated to indicate the progress.
		 *
		 * @param [options] Options object for the task creation.
		 * @param [options.weight] For more information about weight, see the
		 * {@link CKEDITOR.plugins.notificationAggregator.task} overview.
		 * @returns {CKEDITOR.plugins.notificationAggregator.task} An object that represents the task state, and allows
		 * for its manipulation.
		 */
		createTask: function( options ) {
			options = options || {};

			var initialTask = !this.notification,
				task;

			if ( initialTask ) {
				// It's a first call.
				this.notification = this._createNotification();
			}

			task = this._addTask( options );

			task.on( 'updated', this._onTaskUpdate, this );
			task.on( 'done', this._onTaskDone, this );
			task.on( 'canceled', function() {
				this._removeTask( task );
			}, this );

			// Update the aggregator.
			this.update();

			if ( initialTask ) {
				this.notification.show();
			}

			return task;
		},

		/**
		 * Triggers an update on the aggregator, meaning that its UI will be refreshed.
		 *
		 * When all the tasks are done, the {@link #finished} event is fired.
		 */
		update: function() {
			this._updateNotification();

			if ( this.isFinished() ) {
				this.fire( 'finished' );
			}
		},

		/**
		 * Returns a number from `0` to `1` representing the done weights to total weights ratio
		 * (showing how many of the tasks are done).
		 *
		 * Note: For an empty aggregator (without any tasks created) it will return `1`.
		 *
		 * @returns {Number} Returns the percentage of tasks done as a number ranging from `0` to `1`.
		 */
		getPercentage: function() {
			// In case there are no weights at all we'll return 1.
			if ( this.getTaskCount() === 0 ) {
				return 1;
			}

			return this._doneWeights / this._totalWeights;
		},

		/**
		 * @returns {Boolean} Returns `true` if all notification tasks are done
		 * (or there are no tasks at all).
		 */
		isFinished: function() {
			return this.getDoneTaskCount() === this.getTaskCount();
		},

		/**
		 * @returns {Number} Returns a total tasks count.
		 */
		getTaskCount: function() {
			return this._tasks.length;
		},

		/**
		 * @returns {Number} Returns the number of tasks done.
		 */
		getDoneTaskCount: function() {
			return this._doneTasks;
		},

		/**
		 * Updates the notification content.
		 *
		 * @private
		 */
		_updateNotification: function() {
			this.notification.update( {
				message: this._getNotificationMessage(),
				progress: this.getPercentage()
			} );
		},

		/**
		 * Returns a message used in the notification.
		 *
		 * @private
		 * @returns {String}
		 */
		_getNotificationMessage: function() {
			var tasksCount = this.getTaskCount(),
				doneTasks = this.getDoneTaskCount(),
				templateParams = {
					current: doneTasks,
					max: tasksCount,
					percentage: Math.round( this.getPercentage() * 100 )
				},
				template;

			// If there's only one remaining task and we have a singular message, we should use it.
			if ( tasksCount == 1 && this._singularMessage ) {
				template = this._singularMessage;
			} else {
				template = this._message;
			}

			return template.output( templateParams );
		},

		/**
		 * Creates a notification object.
		 *
		 * @private
		 * @returns {CKEDITOR.plugins.notification}
		 */
		_createNotification: function() {
			return new CKEDITOR.plugins.notification( this.editor, {
				type: 'progress'
			} );
		},

		/**
		 * Creates a {@link CKEDITOR.plugins.notificationAggregator.task} instance based
		 * on `options`, and adds it to the task list.
		 *
		 * @private
		 * @param options Options object coming from the {@link #createTask} method.
		 * @returns {CKEDITOR.plugins.notificationAggregator.task}
		 */
		_addTask: function( options ) {
			var task = new Task( options.weight );
			this._tasks.push( task );
			this._totalWeights += task._weight;
			return task;
		},

		/**
		 * Removes a given task from the {@link #_tasks} array and updates the UI.
		 *
		 * @private
		 * @param {CKEDITOR.plugins.notificationAggregator.task} task Task to be removed.
		 */
		_removeTask: function( task ) {
			var index = CKEDITOR.tools.indexOf( this._tasks, task );

			if ( index !== -1 ) {
				// If task was already updated with some weight, we need to remove
				// this weight from our cache.
				if ( task._doneWeight ) {
					this._doneWeights -= task._doneWeight;
				}

				this._totalWeights -= task._weight;

				this._tasks.splice( index, 1 );
				// And we also should inform the UI about this change.
				this.update();
			}
		},

		/**
		 * A listener called on the {@link CKEDITOR.plugins.notificationAggregator.task#update} event.
		 *
		 * @private
		 * @param {CKEDITOR.eventInfo} evt Event object of the {@link CKEDITOR.plugins.notificationAggregator.task#update} event.
		 */
		_onTaskUpdate: function( evt ) {
			this._doneWeights += evt.data;
			this.update();
		},

		/**
		 * A listener called on the {@link CKEDITOR.plugins.notificationAggregator.task#event-done} event.
		 *
		 * @private
		 * @param {CKEDITOR.eventInfo} evt Event object of the {@link CKEDITOR.plugins.notificationAggregator.task#event-done} event.
		 */
		_onTaskDone: function() {
			this._doneTasks += 1;
			this.update();
		}
	};

	CKEDITOR.event.implementOn( Aggregator.prototype );

	/**
	 * # Overview
	 *
	 * This type represents a single task in the aggregator, and exposes methods to manipulate its state.
	 *
	 * ## Weights
	 *
	 * Task progess is based on its **weight**.
	 *
	 * As you create a task, you need to declare its weight. As you want the update to inform about the
	 * progress, you will need to {@link #update} the task, telling how much of this weight is done.
	 *
	 * For example, if you declare that your task has a weight that equals `50` and then call `update` with `10`,
	 * you will end up with telling that the task is done in 20%.
	 *
	 * ### Example Usage of Weights
	 *
	 * Let us say that you use tasks for file uploading.
	 *
	 * A single task is associated with a single file upload. You can use the file size in bytes as a weight,
	 * and then as the file upload progresses you just call the `update` method with the number of bytes actually
	 * downloaded.
	 *
	 * @since 4.5
	 * @class CKEDITOR.plugins.notificationAggregator.task
	 * @mixins CKEDITOR.event
	 * @constructor Creates a task instance for notification aggregator.
	 * @param {Number} [weight=1]
	 */
	function Task( weight ) {
		/**
		 * Total weight of the task.
		 *
		 * @private
		 * @property {Number}
		 */
		this._weight = weight || 1;

		/**
		 * Done weight of the task.
		 *
		 * @private
		 * @property {Number}
		 */
		this._doneWeight = 0;

		/**
		 * Indicates when the task is canceled.
		 *
		 * @private
		 * @property {Boolean}
		 */
		this._isCanceled = false;
	}

	Task.prototype = {
		/**
		 * Marks the task as done.
		 */
		done: function() {
			this.update( this._weight );
		},

		/**
		 * Updates the done weight of a task.
		 *
		 * @param {Number} weight Number indicating how much of the total task {@link #_weight} is done.
		 */
		update: function( weight ) {
			// If task is already done or canceled there is no need to update it, and we don't expect
			// progress to be reversed.
			if ( this.isDone() || this.isCanceled() ) {
				return;
			}

			// Note that newWeight can't be higher than _doneWeight.
			var newWeight = Math.min( this._weight, weight ),
				weightChange = newWeight - this._doneWeight;

			this._doneWeight = newWeight;

			// Fire updated event even if task is done in order to correctly trigger updating the
			// notification's message. If we wouldn't do this, then the last weight change would be ignored.
			this.fire( 'updated', weightChange );

			if ( this.isDone() ) {
				this.fire( 'done' );
			}
		},

		/**
		 * Cancels the task (the task will be removed from the aggregator).
		 */
		cancel: function() {
			// If task is already done or canceled.
			if ( this.isDone() || this.isCanceled() ) {
				return;
			}

			// Mark task as canceled.
			this._isCanceled = true;

			// We'll fire cancel event it's up to aggregator to listen for this event,
			// and remove the task.
			this.fire( 'canceled' );
		},

		/**
		 * Checks if the task is done.
		 *
		 * @returns {Boolean}
		 */
		isDone: function() {
			return this._weight === this._doneWeight;
		},

		/**
		 * Checks if the task is canceled.
		 *
		 * @returns {Boolean}
		 */
		isCanceled: function() {
			return this._isCanceled;
		}
	};

	CKEDITOR.event.implementOn( Task.prototype );

	/**
	 * Fired when all tasks are done. When this event occurs, the notification may change its type to `success` or be hidden:
	 *
	 *		aggregator.on( 'finished', function() {
	 *			if ( aggregator.getTaskCount() == 0 ) {
	 *				aggregator.notification.hide();
	 *			} else {
	 *				aggregator.notification.update( { message: 'Done', type: 'success' } );
	 *			}
	 *		} );
	 *
	 * @event finished
	 * @member CKEDITOR.plugins.notificationAggregator
	 */

	/**
	 * Fired upon each weight update of the task.
	 *
	 *		var myTask = new Task( 100 );
	 *		myTask.update( 30 );
	 *		// Fires updated event with evt.data = 30.
	 *		myTask.update( 40 );
	 *		// Fires updated event with evt.data = 10.
	 *		myTask.update( 20 );
	 *		// Fires updated event with evt.data = -20.
	 *
	 * @event updated
	 * @param {Number} data The difference between the new weight and the previous one.
	 * @member CKEDITOR.plugins.notificationAggregator.task
	 */

	/**
	 * Fired when the task is done.
	 *
	 * @event done
	 * @member CKEDITOR.plugins.notificationAggregator.task
	 */

	/**
	 * Fired when the task is canceled.
	 *
	 * @event canceled
	 * @member CKEDITOR.plugins.notificationAggregator.task
	 */

	// Expose Aggregator type.
	CKEDITOR.plugins.notificationAggregator = Aggregator;
	CKEDITOR.plugins.notificationAggregator.task = Task;
} )();
