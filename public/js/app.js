var TIMELINE_HEIGHT = 200;

var state = 'simple';

// Data
var PPM, // People per minute
    PPH, // People per hour
    data,
    dates,
    maxPPH = 0, // Max PPH
    minPPH = 0, // Min PPH
    totalVisits = 0,
    peopleEntered = 0,
    peopleExited = 0,
    currentVisitors = 0,
    popularHours = [],
    unpopularHours = [],
    dateIndex = 0;

// Timeouts
var fetchTimelineID;

// Processing
var stepX,
    stepY,
    bShouldRedraw = false;

// HTML elements
var simpleElement,
    simpleDateElement,
    timelineElement,
    businessElement,
    popularHoursElement,
    unpopularHoursElement,
    totalVisitsElement,
    currentVisitorsElement,
    enterCountElement,
    exitCountElement,
    timelineDateElement;

function setup() {
	var canvas = createCanvas(windowWidth, 260);
	canvas.parent('canvas-container');
	background(255);
	stroke(245);
	textAlign(CENTER);

	simpleElement = document.getElementById('simple');
	simpleDateElement = document.getElementById('simple-date');
	timelineElement = document.getElementById('timeline');
	businessElement = document.getElementById('business');
	popularHoursElement = document.getElementById('popular-hours');
	unpopularHoursElement = document.getElementById('unpopular-hours');
	totalVisitsElement = document.getElementById('total-visits');
	currentVisitorsElement = document.getElementById('current-visitors');
	enterCountElement = document.getElementById('enter-count');
	exitCountElement = document.getElementById('exit-count');
	timelineDateElement = document.getElementById('timeline-date');

	fetchDates();
	fetchCurrent();

	noLoop();
}

function windowResized() {
	resizeCanvas(windowWidth, 260);
	stroke(245);
	textAlign(CENTER);

	if (state == 'simple') {
		drawSimple();
	} else if (state == 'timeline') {
		drawTimeline();
	}
}

function drawSimple() {
	simpleElement.style.opacity = 1;
	timelineElement.style.opacity = 0;

	if (typeof(PPH) == 'undefined') {
		return;
	}

	var now = new Date();
	var business = maxPPH > 0 ? PPH[now.getHours()] / maxPPH : 0;

	if (business <= 0.2) {
		simpleElement.style.background = '#00ff00';
		businessElement.innerHTML = 'Very Low';
	} else if (business <= 0.4) {
		simpleElement.style.background = '#80ff00';
		businessElement.innerHTML = 'Low';
	} else if (business <= 0.6) {
		simpleElement.style.background = '#ffff00';
		businessElement.innerHTML = 'Medium';
	} else if (business <= 0.8) {
		simpleElement.style.background = '#ff8000';
		businessElement.innerHTML = 'High';
	} else if (business <= 1) {
		simpleElement.style.background = '#ff0000';
		businessElement.innerHTML = 'Very High';
	}
}

function drawTimeline() {
	simpleElement.style.opacity = 0;
	timelineElement.style.opacity = 1;

	if (typeof(PPM) != 'object') {
		return;
	}

	background(255);

	// Calculate stepX
	stepX = (windowWidth * 0.8) / 25;

	push();
	translate(windowWidth * 0.1, 0);

	// Grid
	for (var i = 0; i < 26; i++) {
		line(i * stepX, 0, i * stepX, TIMELINE_HEIGHT); // Row
		line(0, 0, i * stepX, 0); // Top line
		line(0, TIMELINE_HEIGHT, i * stepX, TIMELINE_HEIGHT); // Below line
	}


	// Legend
	fill(50);
	textSize(12);
	var legendText = 'Rate of people entering the store';
	var legendWidth = textWidth(legendText);
	textAlign(LEFT);
	text(legendText, windowWidth * 0.475 - legendWidth - 10, 22);
	textAlign(CENTER);
	fill(185, 235, 223, 200)
	rect(windowWidth * 0.475 - legendWidth - 30, 10, 15, 15);


	// Number of people 
	beginShape();
	fill(185, 235, 223, 200); // Light green
	noStroke();
	for (var i = 0; i < 24; i++) {
		vertex((i + 1) * stepX, TIMELINE_HEIGHT - PPH[i] * stepY); // Shape
	};
	vertex(24 * stepX, 200);
	vertex(stepX, 200);
	endShape(CLOSE);	

	// Dots
	fill(90, 180, 160);
	stroke(255);
	textSize(10);
	for(var i = 0; i < 24; i++) {
		ellipse((i+1) * stepX, TIMELINE_HEIGHT - PPH[i] * stepY, 8, 8);
	}

	// Y-axis
	fill(50);
	text(maxPPH, 0, TIMELINE_HEIGHT * 0.2);
	text(maxPPH/2, 0, TIMELINE_HEIGHT * 0.6);
	text(0, 0, TIMELINE_HEIGHT * 1);

	// Time Bar
	fill(17, 56, 83);
	rect(0, TIMELINE_HEIGHT, 25 * stepX, 55);
	noStroke();
	fill(255);
	textSize(14);
	for(var i = 0; i < 24; i++) {
		var x = (i + 1) * stepX;
		var y = TIMELINE_HEIGHT + 20;
		text(i % 12 + 1, x, y);
		if (i == 0 || i == 23) {
			text('AM', x, y + 20);
		} else if (i == 11) {
			text('PM', x, y + 20);
		}
	}

	pop();
}

function mouseMoved() {
	if (typeof(PPM) != 'object') {
		return;
	}

	for (var i = 0; i < 24; i++) {
		var x = windowWidth * 0.1 + (i + 1) * stepX;
		var y = TIMELINE_HEIGHT - (PPH[i] * stepY);
		if (dist(x, y, mouseX, mouseY) <= 8) {
			fill(90, 180, 160);
			text(PPH[i].toFixed(), x, y - 10);
			bShouldRedraw = true;
			return;
		}
	}

	if (bShouldRedraw) {
		background(255);
		drawTimeline();
		bShouldRedraw = false;
	}
}

function fetchDates() {
	$.ajax({
		url: '/dates',
		method: 'GET',
		dataType: 'json',
	}).done(function(dates_) {
		dates = dates_;
		if (!dates || dates.length == 0) {
			return;
		}
		console.log('Loaded dates');

		dateIndex = 0;
		fetchTimeline();
	}).fail(function(resp) {
		alert('Failed to fetch dates!');
	});
}

function fetchTimeline(direction) {
	if (dates.length == 0) {
		console.log('No dates');
		return;
	}

	if (typeof(direction) == 'undefined') {
		direction = 0;
	}

	var tmpIndex = dateIndex + direction;
	if (tmpIndex < 0 || tmpIndex >= dates.length) {
		tmpIndex = dateIndex;
	}

	var date = dates[tmpIndex];
	$.ajax({
		url: '/timeline',
		method: 'GET',
		data: { year: date.year, month: date.month, day: date.day },
		dataType: 'json',
	}).done(function(data_) {
		data = data_;
		dateIndex = tmpIndex;
		maxPPH = 0;
		minPPH = 99999999;
		console.log('Loaded timeline');

		// Convert data to a time-specific format
		PPM = [];
		for (var i in data_) {
			var count = 0;
			if (i == 0) {
				count = Math.max(data_[i][0], 0);
			} else {
				count = Math.max(data_[i][0] - data_[i - 1][0], 0);
			}
			PPM[i] = count;
		}

		// Count PPH, max and min PPH
		PPH = new Array(24);
		for (var hour = 0; hour < 24; hour++) {
			var tmpCount = 0;
			for (var minute = 0; minute < 60; minute++) {
				var index = hour * 60 + minute;
				if (PPM[index]) {
					tmpCount += PPM[index];
				}
			}

			PPH[hour] = tmpCount;
			maxPPH = Math.max(maxPPH, tmpCount);
			minPPH = Math.min(minPPH, tmpCount);
		}

		// Query popular and unpopular hour
		if (maxPPH > 0) {
			// Get popular hour ranges
			var range = [];
			var prev = -1;
			for (var hour = 0; hour < 24; hour++) {
				if (PPH[hour] >= maxPPH * 0.8) {
					if (prev == hour - 1) {
						range.push(hour);
					} else {
						if (range.length > 0) {
							popularHours.push(range);
						}
						range = [];
						range.push(hour);
					}
					prev = hour;
				}

				if (hour == 23 && range.length > 0) {
					popularHours.push(range);
				}
			}
			var maxHoursPPH = 0, maxHoursPPHRange;
			for (var i = 0; i < popularHours.length; i++) {
				var hours = popularHours[i];
				var hoursPPH = 0;
				for (var j = 0; j < hours.length; j++) {
					var index = hours[j];
					hoursPPH += PPH[index];
				}
				if (maxHoursPPH < hoursPPH) {
					maxHoursPPH = hoursPPH;
					maxHoursPPHRange = popularHours[i];
				}
			}
			if (maxHoursPPHRange) {
				if (maxHoursPPHRange.length > 1) {
					var start = maxHoursPPHRange[0];
					var startSuffix = start < 12 ? 'AM' : 'PM';
					var end = maxHoursPPHRange[maxHoursPPHRange.length - 1];
					var endSuffix = end < 12 ? 'AM' : 'PM';
					popularHoursElement.innerHTML = ((start % 12) + 1) + startSuffix + ' - ' + end + endSuffix;
				} else if (maxHoursPPHRange.length == 1) {
					var start = maxHoursPPHRange[0];
					var startSuffix = start < 12 ? 'AM' : 'PM';
					popularHoursElement.innerHTML = ((start % 12) + 1) + startSuffix;
				} else {
					popularHoursElement.innerHTML = 'N/A';
				}
			}

			// Get unpopular hour ranges
			range = [];
			prev = -1;
			for (var i = 0; i < 24; i++) {
				if (PPH[i] > 0 && PPH[i] <= maxPPH * 0.1) {
					if (prev == i - 1) {
						range.push(i);
					} else {
						if (range.length > 0) {
							unpopularHours.push(range);
						}
						range = [];
						range.push(i);
					}
					prev = i;
				}

				if (i == 23 && range.length > 0) {
					unpopularHours.push(range);
				}
			}

			var minHoursPPH = 99999999, minHoursPPHRange;
			for (var i = 0; i < unpopularHours.length; i++) {
				var hours = unpopularHours[i];
				var hoursPPH = 0;
				for (var j = 0; j < hours.length; j++) {
					var index = hours[j];
					hoursPPH += PPH[index];
				}
				if (minHoursPPH > hoursPPH) {
					minHoursPPH = hoursPPH;
					minHoursPPHRange = unpopularHours[i];
				}
			}
			if (minHoursPPHRange) {
				if (minHoursPPHRange.length > 1) {
					var start = minHoursPPHRange[0];
					var startSuffix = start < 12 ? 'AM' : 'PM';
					var end = minHoursPPHRange[minHoursPPHRange.length - 1];
					var endSuffix = end < 12 ? 'AM' : 'PM';
					unpopularHoursElement.innerHTML = ((start % 12) + 1) + startSuffix + ' - ' + end + endSuffix;
				} else if (minHoursPPHRange.length == 1) {
					var start = minHoursPPHRange[0];
					var startSuffix = start < 12 ? 'AM' : 'PM';
					unpopularHoursElement.innerHTML = ((start % 12) + 1) + startSuffix;
				} else {
					unpopularHoursElement.innerHTML = 'N/A';
				}
			}
		}

		// Count total visits
		totalVisits = 0;
		for (var i = 0; i < 24; i++) {
			totalVisits += PPH[i]
		}
		totalVisitsElement.innerHTML = totalVisits;

		// Update stepY
		stepY = maxPPH > 0 ? TIMELINE_HEIGHT * 0.8 / maxPPH : 0;

		// Get JavaScript Date
		date = new Date(date.year, date.month - 1, date.day);

		// Update day element if necessary
		timelineDateElement.innerHTML = date.toDateString();
		simpleDateElement.innerHTML = date.toDateString() + " " + new Date().toTimeString();

		// Redraw
		if (state == 'simple') {
			drawSimple();
		} else if (state == 'timeline') {
			drawTimeline();
		}

		fetchTimelineID = setTimeout(fetchTimeline, 60000);
	}).fail(function(response) {
		fetchTimelineID = setTimeout(fetchTimeline, 60000);
	});
}

function fetchCurrent() {
	$.ajax({
		url: '/visit',
		method: 'GET',
		dataType: 'json',
	}).done(function(data) {
		currentVisitors = Math.max(data.enters - data.exits, 0);
		currentVisitorsElement.innerHTML = currentVisitors;
		enterCountElement.innerHTML = data.enters;
		exitCountElement.innerHTML = data.exits;
		if (data.enters == 0) {
			data.enters = 1;
		}
		enterCountElement.style.width = '100%';
		currentVisitorsElement.style.width = Math.min(currentVisitors / data.enters, 1) * enterCountElement.offsetWidth + 'px';
		exitCountElement.style.width = Math.min(data.exits / data.enters, 1) * enterCountElement.offsetWidth + 'px';
		fetchCurrentID = window.setTimeout(fetchCurrent, 1000);
	}).fail(function(response) {
		fetchCurrentID = window.setTimeout(fetchCurrent, 1000);
	});
}

function keyReleased() {
	if (keyCode == LEFT_ARROW) {
		fetchTimeline(-1);
	}

	if (keyCode == RIGHT_ARROW) {
		fetchTimeline(1);
	}

	if (keyCode == TAB) {
		if (state == 'timeline') {
			state = 'simple';
			drawSimple();
		} else {
			state = 'timeline';
			drawTimeline();
		}
	}
}
