package main

import (
	"encoding/json"

	"github.com/davecgh/go-spew/spew"
	log "github.com/sirupsen/logrus"
)

/*
type Item struct {
	ID       string `json:"id"`
	Abstract string `json:"abstract"`
	Type     string `json:"type"`
	CopyFrom string `json:"copy-from"`
	Raw      json.RawMessage
}
*/

func DoIt(sources map[string][]byte) error {
	templates := make(map[string]map[string]interface{})
	for _, t := range sources {
		var data []map[string]interface{}
		err := json.Unmarshal(t, &data)
		if err != nil {
			log.Fatal(err)
		}

		for _, d := range data {
			if keyExists(d, "abstract") {
				templates[d["abstract"].(string)] = d
			} else if keyExists(d, "id") {
				templates[d["id"].(string)] = d
			} else {
				log.Fatal("no id or abstract")
			}
		}
	}

	for id, t := range templates {
		itemType := t["type"].(string)
		switch itemType {
		case "TOOLMOD":
			buildToolmod(id, templates)
		}
	}
	spew.Dump(templates["mod_battery"])

	return nil
}

func buildToolmod(id string, templates map[string]map[string]interface{}) {

	/*
		void Item_factory::load_toolmod( JsonObject &jo, const std::string &src )
		{
		    itype def;
		    if( load_definition( jo, src, def ) ) {
		        load_slot( def.mod, jo, src );
		        load_basic_info( jo, def, src );
		    }
		}

		bool Item_factory::load_definition( JsonObject& jo, const std::string &src, itype &def ) {
		    assert( !frozen );

		    if( !jo.has_string( "copy-from" ) ) {
		        // if this is a new definition ensure we start with a clean itype
		        def = itype();

		        // adjust type specific defaults
		        auto opt = jo.get_string( "type" );

		        // ammo and comestibles by default lack differing damage levels and are always stackable
		        if( opt == "AMMO" || opt == "COMESTIBLE" ) {
		            def.damage_min = 0;
		            def.damage_max = 0;
		            def.stackable = true;
		        }
		        return true;
		    }

		    auto base = m_templates.find( jo.get_string( "copy-from" ) );
		    if( base != m_templates.end() ) {
		        def = base->second;
		        return true;
		    }

		    auto abstract = m_abstracts.find( jo.get_string( "copy-from" ) );
		    if( abstract != m_abstracts.end() ) {
		        def= abstract->second;
		        return true;
		    }

		    deferred.emplace_back( jo.str(), src );
		    return false;
		}

		void Item_factory::load( islot_mod &slot, JsonObject &jo, const std::string &src )
		{
		    bool strict = src == "dda";

		    assign( jo, "ammo_modifier", slot.ammo_modifier, strict );
		    assign( jo, "capacity_multiplier", slot.capacity_multiplier, strict );

		    if( jo.has_member( "acceptable_ammo" ) ) {
		        slot.acceptable_ammo.clear();
		        for( auto &e : jo.get_tags( "acceptable_ammo" ) ) {
		            slot.acceptable_ammo.insert( ammotype( e ) );
		        }
		    }

		    JsonArray mags = jo.get_array( "magazine_adaptor" );
		    if( !mags.empty() ) {
		        slot.magazine_adaptor.clear();
		    }
		    while( mags.has_more() ) {
		        JsonArray arr = mags.next_array();

		        ammotype ammo( arr.get_string( 0 ) ); // an ammo type (e.g. 9mm)
		        JsonArray compat = arr.get_array( 1 ); // compatible magazines for this ammo type

		        while( compat.has_more() ) {
		            slot.magazine_adaptor[ ammo ].insert( compat.next_string() );
		        }
		    }
		}

		void Item_factory::load_basic_info( JsonObject &jo, itype &def, const std::string &src )
		{
		    bool strict = src == "dda";

		    assign( jo, "category", def.category_force, strict );
		    assign( jo, "weight", def.weight, strict, 0 );
		    assign( jo, "volume", def.volume );
		    assign( jo, "price", def.price );
		    assign( jo, "price_postapoc", def.price_post );
		    assign( jo, "stackable", def.stackable, strict );
		    assign( jo, "integral_volume", def.integral_volume );
		    assign( jo, "bashing", def.melee[DT_BASH], strict, 0 );
		    assign( jo, "cutting", def.melee[DT_CUT], strict, 0 );
		    assign( jo, "to_hit", def.m_to_hit, strict );
		    assign( jo, "container", def.default_container );
		    assign( jo, "rigid", def.rigid );
		    assign( jo, "min_strength", def.min_str );
		    assign( jo, "min_dexterity", def.min_dex );
		    assign( jo, "min_intelligence", def.min_int );
		    assign( jo, "min_perception", def.min_per );
		    assign( jo, "emits", def.emits );
		    assign( jo, "magazine_well", def.magazine_well );
		    assign( jo, "explode_in_fire", def.explode_in_fire );

		    if( jo.has_member( "thrown_damage" ) ) {
		        JsonArray jarr = jo.get_array( "thrown_damage" );
		        def.thrown_damage = load_damage_instance( jarr );
		    } else {
		        // @todo: Move to finalization
		        def.thrown_damage.clear();
		        def.thrown_damage.add_damage( DT_BASH, def.melee[DT_BASH] + def.weight / 1.0_kilogram );
		    }

		    if( jo.has_member( "damage_states" ) ) {
		        auto arr = jo.get_array( "damage_states" );
		        def.damage_min = arr.get_int( 0 );
		        def.damage_max = arr.get_int( 1 );
		    }

		    def.name = jo.get_string( "name" );
		    if( jo.has_member( "name_plural" ) ) {
		        def.name_plural = jo.get_string( "name_plural" );
		    } else {
		        def.name_plural = jo.get_string( "name" ) += "s";
		    }

		    if( jo.has_string( "description" ) ) {
		        def.description = jo.get_string( "description" );
		    }

		    if( jo.has_string( "symbol" ) ) {
		        def.sym = jo.get_string( "symbol" );
		    }

		    if( jo.has_string( "color" ) ) {
		        def.color = color_from_string( jo.get_string( "color" ) );
		    }

		    if( jo.has_member( "material" ) ) {
		        def.materials.clear();
		        for( auto &m : jo.get_tags( "material" ) ) {
		            def.materials.emplace_back( m );
		        }
		    }

		    if( jo.has_string( "phase" ) ) {
		        def.phase = jo.get_enum_value<phase_id>( "phase" );
		    }

		    if( jo.has_array( "magazines" ) ) {
		        def.magazine_default.clear();
		        def.magazines.clear();
		    }
		    JsonArray mags = jo.get_array( "magazines" );
		    while( mags.has_more() ) {
		        JsonArray arr = mags.next_array();

		        ammotype ammo( arr.get_string( 0 ) ); // an ammo type (e.g. 9mm)
		        JsonArray compat = arr.get_array( 1 ); // compatible magazines for this ammo type

		        // the first magazine for this ammo type is the default;
		        def.magazine_default[ ammo ] = compat.get_string( 0 );

		        while( compat.has_more() ) {
		            def.magazines[ ammo ].insert( compat.next_string() );
		        }
		    }

		    JsonArray jarr = jo.get_array( "min_skills" );
		    if( !jarr.empty() ) {
		        def.min_skills.clear();
		    }
		    while( jarr.has_more() ) {
		        JsonArray cur = jarr.next_array();
		        const auto sk = skill_id( cur.get_string( 0 ) );
		        if( !sk.is_valid() ) {
		            jo.throw_error( string_format( "invalid skill: %s", sk.c_str() ), "min_skills" );
		        }
		        def.min_skills[ sk ] = cur.get_int( 1 );
		    }

		    if( jo.has_member("explosion" ) ) {
		        JsonObject je = jo.get_object( "explosion" );
		        def.explosion = load_explosion_data( je );
		    }

		    assign( jo, "flags", def.item_tags );

		    if( jo.has_member( "qualities" ) ) {
		        set_qualities_from_json( jo, "qualities", def );
		    }

		    if( jo.has_member( "properties" ) ) {
		        set_properties_from_json( jo, "properties", def );
		    }

		    for( auto & s : jo.get_tags( "techniques" ) ) {
		        def.techniques.insert( matec_id( s ) );
		    }

		    set_use_methods_from_json( jo, "use_action", def.use_methods );

		    assign( jo, "countdown_interval", def.countdown_interval );
		    assign( jo, "countdown_destroy", def.countdown_destroy );

		    if( jo.has_string( "countdown_action" ) ) {
		        def.countdown_action = usage_from_string( jo.get_string( "countdown_action" ) );

		    } else if( jo.has_object( "countdown_action" ) ) {
		        auto tmp = jo.get_object( "countdown_action" );
		        def.countdown_action = usage_from_object( tmp ).second;
		    }

		    if( jo.has_string( "drop_action" ) ) {
		        def.drop_action = usage_from_string( jo.get_string( "drop_action" ) );

		    } else if( jo.has_object( "drop_action" ) ) {
		        auto tmp = jo.get_object( "drop_action" );
		        def.drop_action = usage_from_object( tmp ).second;
		    }

		    load_slot_optional( def.container, jo, "container_data", src );
		    load_slot_optional( def.armor, jo, "armor_data", src );
		    load_slot_optional( def.book, jo, "book_data", src );
		    load_slot_optional( def.gun, jo, "gun_data", src );
		    load_slot_optional( def.bionic, jo, "bionic_data", src );
		    load_slot_optional( def.ammo, jo, "ammo_data", src );
		    load_slot_optional( def.seed, jo, "seed_data", src );
		    load_slot_optional( def.artifact, jo, "artifact_data", src );
		    load_slot_optional( def.brewable, jo, "brewable", src );
		    load_slot_optional( def.fuel, jo, "fuel", src );

		    // optional gunmod slot may also specify mod data
		    load_slot_optional( def.gunmod, jo, "gunmod_data", src );
		    load_slot_optional( def.mod, jo, "gunmod_data", src );

		    if( jo.has_string( "abstract" ) ) {
		        def.id = jo.get_string( "abstract" );
		    } else {
		        def.id = jo.get_string( "id" );
		    }

		    // snippet_category should be loaded after def.id is determined
		    if( jo.has_array( "snippet_category" ) ) {
		        // auto-create a category that is unlikely to already be used and put the
		        // snippets in it.
		        def.snippet_category = std::string( "auto:" ) + def.id;
		        JsonArray jarr = jo.get_array( "snippet_category" );
		        SNIPPET.add_snippets_from_json( def.snippet_category, jarr );
		    } else {
		        def.snippet_category = jo.get_string( "snippet_category", "" );
		    }

		    if( jo.has_string( "abstract" ) ) {
		        m_abstracts[ def.id ] = def;
		    } else {
		        m_templates[ def.id ] = def;
		    }
		}
	*/

	template := templates[id]

	spew.Dump(template)
}

func keyExists(decoded map[string]interface{}, key string) bool {
	val, ok := decoded[key]
	return ok && val != nil
}
