var date = new Date().toUTCString();
console.log('Getting stats for date:', date);

$.ajax({
	url: '/stats',
	method: 'GET',
	data: { date: date },
}).done(function(stats) {
	console.log('Got stats:', stats);
}).fail(function(resp) {
	console.log('Failed to get stats');
});
