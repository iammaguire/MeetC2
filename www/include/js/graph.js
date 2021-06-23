var graph = new Springy.Graph();

// make some nodes
var spruce = graph.newNode({label: 'Norway Spruce'});
var fir = graph.newNode({label: 'Sicilian Fir'});

// connect them with an edge
graph.newEdge(spruce, fir);

var layout = new Springy.Layout.ForceDirected(
    graph,
    400.0, // Spring stiffness
    400.0, // Node repulsion
    0.5 // Damping
  );

  renderer.start();


  $('#graph').springy({ graph: graph });