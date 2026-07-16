<?xml version="1.0" encoding="UTF-8"?>
<xsl:stylesheet xmlns:xsl="http://www.w3.org/1999/XSL/Transform"
                xmlns:math="http://exslt.org/math"
                xmlns:exslt="http://exslt.org/common"
                xmlns:western="urn:empirical:western"
                version="1.0"
                exclude-result-prefixes="math exslt">

  <!-- ── Parameters ────────────────────────────────────────────────────── -->
  <xsl:param name="orb" select="8.0"/>

  <!-- ── Western planets (19: classical + modern + asteroids + dwarfs) ── -->
  <!-- Excludes Node/TrueNode (handled in Nodes section) and TNPs (Uranian). -->
  <xsl:variable name="westernPlanets" select="',Sun,Moon,Mercury,Venus,Mars,Jupiter,Saturn,Uranus,Neptune,Pluto,Ceres,Pallas,Juno,Vesta,Lilith,Chiron,Eris,Makemake,Gonggong,'"/>

  <!-- ── Modern aspects (9 types) ──────────────────────────────────────── -->
  <xsl:variable name="aspects">
    <aspect angle="0" name="conjunction"/>
    <aspect angle="30" name="semi-sextile"/>
    <aspect angle="45" name="semi-square"/>
    <aspect angle="60" name="sextile"/>
    <aspect angle="90" name="square"/>
    <aspect angle="120" name="trine"/>
    <aspect angle="135" name="sesquiquadrate"/>
    <aspect angle="150" name="quincunx"/>
    <aspect angle="180" name="opposition"/>
  </xsl:variable>

  <!-- ── Sign names ────────────────────────────────────────────────────── -->
  <xsl:variable name="signs" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>

  <!-- ── Domicile rulers (planet → sign indices, 0-based) ──────────────── -->
  <!-- Traditional + modern: Uranus→Aquarius, Neptune→Pisces, Pluto→Scorpio -->
  <xsl:variable name="domicileSun" select="4"/>
  <xsl:variable name="domicileMoon" select="3"/>
  <xsl:variable name="domicileMercury1" select="2"/>
  <xsl:variable name="domicileMercury2" select="5"/>
  <xsl:variable name="domicileVenus1" select="1"/>
  <xsl:variable name="domicileVenus2" select="6"/>
  <xsl:variable name="domicileMars1" select="0"/>
  <xsl:variable name="domicileMars2" select="7"/>
  <xsl:variable name="domicileJupiter1" select="8"/>
  <xsl:variable name="domicileJupiter2" select="11"/>
  <xsl:variable name="domicileSaturn1" select="9"/>
  <xsl:variable name="domicileSaturn2" select="10"/>
  <xsl:variable name="domicileUranus" select="10"/>   <!-- Aquarius -->
  <xsl:variable name="domicileNeptune" select="11"/>  <!-- Pisces -->
  <xsl:variable name="domicilePluto" select="7"/>     <!-- Scorpio -->

  <!-- ── Exaltation (planet → sign index) ──────────────────────────────── -->
  <xsl:variable name="exaltationSun" select="0"/>      <!-- Aries -->
  <xsl:variable name="exaltationMoon" select="1"/>     <!-- Taurus -->
  <xsl:variable name="exaltationMercury" select="5"/>  <!-- Virgo -->
  <xsl:variable name="exaltationVenus" select="11"/>    <!-- Pisces -->
  <xsl:variable name="exaltationMars" select="9"/>     <!-- Capricorn -->
  <xsl:variable name="exaltationJupiter" select="3"/>  <!-- Cancer -->
  <xsl:variable name="exaltationSaturn" select="6"/>   <!-- Libra -->
  <!-- Modern exaltations (less consensus, but common assignments) -->
  <xsl:variable name="exaltationUranus" select="7"/>   <!-- Scorpio -->
  <xsl:variable name="exaltationNeptune" select="3"/>  <!-- Cancer (some say Leo) -->
  <xsl:variable name="exaltationPluto" select="0"/>    <!-- Aries (some say Leo) -->

  <!-- ── Fall (opposite exaltation, sign index) ────────────────────────── -->
  <xsl:variable name="fallSun" select="6"/>            <!-- Libra -->
  <xsl:variable name="fallMoon" select="7"/>           <!-- Scorpio -->
  <xsl:variable name="fallMercury" select="11"/>       <!-- Pisces -->
  <xsl:variable name="fallVenus" select="5"/>          <!-- Virgo -->
  <xsl:variable name="fallMars" select="3"/>           <!-- Cancer -->
  <xsl:variable name="fallJupiter" select="9"/>        <!-- Capricorn -->
  <xsl:variable name="fallSaturn" select="0"/>         <!-- Aries -->
  <xsl:variable name="fallUranus" select="1"/>         <!-- Taurus -->
  <xsl:variable name="fallNeptune" select="9"/>        <!-- Capricorn -->
  <xsl:variable name="fallPluto" select="6"/>          <!-- Libra -->

  <!-- ── Root template ────────────────────────────────────────────────── -->
  <xsl:template match="/BaseChart">
    <western:Chart xmlns:western="urn:empirical:western">
      <western:Name><xsl:value-of select="Identity/Name"/></western:Name>

      <!-- Determine sect -->
      <xsl:variable name="sunLon" select="Positions/Planet[@name='Sun']/Tropical/Lon"/>
      <xsl:variable name="ascLon" select="Angles/ASC"/>
      <xsl:variable name="sunAscDiff" select="($sunLon - $ascLon + 360) mod 360"/>
      <xsl:variable name="isDay" select="$sunAscDiff &lt; 180"/>

      <western:Sect>
        <xsl:choose>
          <xsl:when test="$isDay">day</xsl:when>
          <xsl:otherwise>night</xsl:otherwise>
        </xsl:choose>
      </western:Sect>

      <!-- Planet signs -->
      <western:PlanetSigns>
        <xsl:for-each select="Positions/Planet[contains($westernPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <western:PlanetSign>
            <western:Planet><xsl:value-of select="@name"/></western:Planet>
            <western:Sign>
              <xsl:call-template name="lonToSign">
                <xsl:with-param name="lon" select="Tropical/Lon"/>
              </xsl:call-template>
            </western:Sign>
            <western:Degree>
              <xsl:value-of select="format-number(Tropical/Lon mod 30, '0.00')"/>
            </western:Degree>
            <western:Retrograde>
              <xsl:value-of select="Tropical/Speed &lt; 0"/>
            </western:Retrograde>
          </western:PlanetSign>
        </xsl:for-each>
      </western:PlanetSigns>

      <!-- Placidus houses -->
      <western:PlanetHouses>
        <xsl:for-each select="Positions/Planet[contains($westernPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <western:PlanetHouse>
            <western:Planet><xsl:value-of select="@name"/></western:Planet>
            <western:House>
              <xsl:call-template name="lonToPlacidusHouse">
                <xsl:with-param name="lon" select="Tropical/Lon"/>
              </xsl:call-template>
            </western:House>
          </western:PlanetHouse>
        </xsl:for-each>
      </western:PlanetHouses>

      <!-- Aspects (9 types, modern planets) -->
      <western:Aspects>
        <xsl:call-template name="computeAspects">
          <xsl:with-param name="planets" select="Positions/Planet[contains($westernPlanets, concat(',', @name, ','))]"/>
        </xsl:call-template>
      </western:Aspects>

      <!-- Essential dignities (domicile/detriment/exaltation/fall) -->
      <western:Dignities>
        <xsl:for-each select="Positions/Planet[contains($westernPlanets, concat(',', @name, ','))]">
          <xsl:sort select="@id" data-type="number"/>
          <xsl:call-template name="assessDignity">
            <xsl:with-param name="planet" select="."/>
          </xsl:call-template>
        </xsl:for-each>
      </western:Dignities>

      <!-- Nodes -->
      <western:Nodes>
        <western:NorthNode>
          <western:Sign>
            <xsl:call-template name="lonToSign">
              <xsl:with-param name="lon" select="Nodes/NorthNode"/>
            </xsl:call-template>
          </western:Sign>
          <western:Degree>
            <xsl:value-of select="format-number(Nodes/NorthNode mod 30, '0.00')"/>
          </western:Degree>
        </western:NorthNode>
        <western:SouthNode>
          <western:Sign>
            <xsl:call-template name="lonToSign">
              <xsl:with-param name="lon" select="Nodes/SouthNode"/>
            </xsl:call-template>
          </western:Sign>
          <western:Degree>
            <xsl:value-of select="format-number(Nodes/SouthNode mod 30, '0.00')"/>
          </western:Degree>
        </western:SouthNode>
      </western:Nodes>

      <!-- Angles -->
      <western:Angles>
        <western:ASC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/ASC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/ASC mod 30, '0.00')"/>
        </western:ASC>
        <western:MC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/MC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/MC mod 30, '0.00')"/>
        </western:MC>
        <western:DSC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/DSC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/DSC mod 30, '0.00')"/>
        </western:DSC>
        <western:IC>
          <xsl:call-template name="lonToSign">
            <xsl:with-param name="lon" select="Angles/IC"/>
          </xsl:call-template>
          <xsl:text> </xsl:text>
          <xsl:value-of select="format-number(Angles/IC mod 30, '0.00')"/>
        </western:IC>
      </western:Angles>

    </western:Chart>
  </xsl:template>

  <!-- ══════════════════════════════════════════════════════════════════════
       HELPERS
       ════════════════════════════════════════════════════════════════════ -->

  <!-- ── Longitude to sign name ────────────────────────────────────────── -->
  <xsl:template name="lonToSign">
    <xsl:param name="lon"/>
    <xsl:variable name="idx" select="floor($lon div 30)"/>
    <xsl:call-template name="nthToken">
      <xsl:with-param name="list" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>
      <xsl:with-param name="n" select="$idx + 1"/>
    </xsl:call-template>
  </xsl:template>

  <!-- ── nth comma-separated token ─────────────────────────────────────── -->
  <xsl:template name="nthToken">
    <xsl:param name="list"/>
    <xsl:param name="n"/>
    <xsl:choose>
      <xsl:when test="$n = 1">
        <xsl:value-of select="substring-before(concat($list, ','), ',')"/>
      </xsl:when>
      <xsl:otherwise>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="substring-after($list, ',')"/>
          <xsl:with-param name="n" select="$n - 1"/>
        </xsl:call-template>
      </xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Longitude to Placidus house ───────────────────────────────────── -->
  <!-- Uses absolute paths because this template is called from within
       for-each over Planet nodes, so relative paths would fail. -->
  <xsl:template name="lonToPlacidusHouse">
    <xsl:param name="lon"/>
    <xsl:variable name="c1" select="/BaseChart/Houses/System[@name='placidus']/Cusp[1]"/>
    <xsl:variable name="c2" select="/BaseChart/Houses/System[@name='placidus']/Cusp[2]"/>
    <xsl:variable name="c3" select="/BaseChart/Houses/System[@name='placidus']/Cusp[3]"/>
    <xsl:variable name="c4" select="/BaseChart/Houses/System[@name='placidus']/Cusp[4]"/>
    <xsl:variable name="c5" select="/BaseChart/Houses/System[@name='placidus']/Cusp[5]"/>
    <xsl:variable name="c6" select="/BaseChart/Houses/System[@name='placidus']/Cusp[6]"/>
    <xsl:variable name="c7" select="/BaseChart/Houses/System[@name='placidus']/Cusp[7]"/>
    <xsl:variable name="c8" select="/BaseChart/Houses/System[@name='placidus']/Cusp[8]"/>
    <xsl:variable name="c9" select="/BaseChart/Houses/System[@name='placidus']/Cusp[9]"/>
    <xsl:variable name="c10" select="/BaseChart/Houses/System[@name='placidus']/Cusp[10]"/>
    <xsl:variable name="c11" select="/BaseChart/Houses/System[@name='placidus']/Cusp[11]"/>
    <xsl:variable name="c12" select="/BaseChart/Houses/System[@name='placidus']/Cusp[12]"/>

    <!-- Placidus cusps can wrap around 360° (e.g. cusp 6 at 9°, cusp 7 at 30°).
         Normalize: if cusp[n] > cusp[n+1], the house spans the 0° boundary.
         Strategy: shift all longitudes so cusp 1 (ASC) is at 0, then compare. -->
    <xsl:variable name="shift" select="360 - $c1"/>
    <xsl:variable name="normLon" select="($lon + $shift) mod 360"/>
    <xsl:variable name="n1" select="0"/>
    <xsl:variable name="n2" select="($c2 + $shift) mod 360"/>
    <xsl:variable name="n3" select="($c3 + $shift) mod 360"/>
    <xsl:variable name="n4" select="($c4 + $shift) mod 360"/>
    <xsl:variable name="n5" select="($c5 + $shift) mod 360"/>
    <xsl:variable name="n6" select="($c6 + $shift) mod 360"/>
    <xsl:variable name="n7" select="($c7 + $shift) mod 360"/>
    <xsl:variable name="n8" select="($c8 + $shift) mod 360"/>
    <xsl:variable name="n9" select="($c9 + $shift) mod 360"/>
    <xsl:variable name="n10" select="($c10 + $shift) mod 360"/>
    <xsl:variable name="n11" select="($c11 + $shift) mod 360"/>
    <xsl:variable name="n12" select="($c12 + $shift) mod 360"/>

    <xsl:choose>
      <xsl:when test="$normLon >= $n1 and $normLon &lt; $n2">1</xsl:when>
      <xsl:when test="$normLon >= $n2 and $normLon &lt; $n3">2</xsl:when>
      <xsl:when test="$normLon >= $n3 and $normLon &lt; $n4">3</xsl:when>
      <xsl:when test="$normLon >= $n4 and $normLon &lt; $n5">4</xsl:when>
      <xsl:when test="$normLon >= $n5 and $normLon &lt; $n6">5</xsl:when>
      <xsl:when test="$normLon >= $n6 and $normLon &lt; $n7">6</xsl:when>
      <xsl:when test="$normLon >= $n7 and $normLon &lt; $n8">7</xsl:when>
      <xsl:when test="$normLon >= $n8 and $normLon &lt; $n9">8</xsl:when>
      <xsl:when test="$normLon >= $n9 and $normLon &lt; $n10">9</xsl:when>
      <xsl:when test="$normLon >= $n10 and $normLon &lt; $n11">10</xsl:when>
      <xsl:when test="$normLon >= $n11 and $normLon &lt; $n12">11</xsl:when>
      <xsl:otherwise>12</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Compute aspects between planet set ────────────────────────────── -->
  <xsl:template name="computeAspects">
    <xsl:param name="planets"/>
    <xsl:for-each select="$planets">
      <xsl:variable name="p1pos" select="position()"/>
      <xsl:variable name="p1" select="."/>
      <xsl:variable name="p1lon" select="$p1/Tropical/Lon"/>
      <xsl:variable name="p1name" select="$p1/@name"/>
      <xsl:for-each select="$planets[position() > $p1pos]">
        <xsl:variable name="p2lon" select="Tropical/Lon"/>
        <xsl:variable name="p2name" select="@name"/>
        <xsl:variable name="dist" select="($p2lon - $p1lon + 360) mod 360"/>
        <!-- Check all 9 aspect angles -->
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="0"/>
          <xsl:with-param name="aspectName" select="'conjunction'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="30"/>
          <xsl:with-param name="aspectName" select="'semi-sextile'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="45"/>
          <xsl:with-param name="aspectName" select="'semi-square'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="60"/>
          <xsl:with-param name="aspectName" select="'sextile'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="90"/>
          <xsl:with-param name="aspectName" select="'square'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="120"/>
          <xsl:with-param name="aspectName" select="'trine'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="135"/>
          <xsl:with-param name="aspectName" select="'sesquiquadrate'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="150"/>
          <xsl:with-param name="aspectName" select="'quincunx'"/>
        </xsl:call-template>
        <xsl:call-template name="checkAspect">
          <xsl:with-param name="p1name" select="$p1name"/>
          <xsl:with-param name="p2name" select="$p2name"/>
          <xsl:with-param name="dist" select="$dist"/>
          <xsl:with-param name="angle" select="180"/>
          <xsl:with-param name="aspectName" select="'opposition'"/>
        </xsl:call-template>
      </xsl:for-each>
    </xsl:for-each>
  </xsl:template>

  <!-- ── Check a single aspect angle ──────────────────────────────────── -->
  <xsl:template name="checkAspect">
    <xsl:param name="p1name"/>
    <xsl:param name="p2name"/>
    <xsl:param name="dist"/>
    <xsl:param name="angle"/>
    <xsl:param name="aspectName"/>
    <xsl:variable name="diff" select="$dist - $angle"/>
    <xsl:variable name="absDiff">
      <xsl:choose>
        <xsl:when test="$diff &lt; 0"><xsl:value-of select="-$diff"/></xsl:when>
        <xsl:otherwise><xsl:value-of select="$diff"/></xsl:otherwise>
      </xsl:choose>
    </xsl:variable>
    <xsl:variable name="oppDiff" select="360 - $dist - $angle"/>
    <xsl:variable name="absOppDiff">
      <xsl:choose>
        <xsl:when test="$oppDiff &lt; 0"><xsl:value-of select="-$oppDiff"/></xsl:when>
        <xsl:otherwise><xsl:value-of select="$oppDiff"/></xsl:otherwise>
      </xsl:choose>
    </xsl:variable>
    <xsl:if test="$absDiff &lt;= $orb or $absOppDiff &lt;= $orb">
      <western:Aspect>
        <western:Planet1><xsl:value-of select="$p1name"/></western:Planet1>
        <western:Planet2><xsl:value-of select="$p2name"/></western:Planet2>
        <western:Type><xsl:value-of select="$aspectName"/></western:Type>
        <western:Orb>
          <xsl:choose>
            <xsl:when test="$absDiff &lt;= $absOppDiff">
              <xsl:value-of select="format-number($absDiff, '0.00')"/>
            </xsl:when>
            <xsl:otherwise>
              <xsl:value-of select="format-number($absOppDiff, '0.00')"/>
            </xsl:otherwise>
          </xsl:choose>
        </western:Orb>
      </western:Aspect>
    </xsl:if>
  </xsl:template>

  <!-- ── Assess essential dignity (domicile/detriment/exaltation/fall) ── -->
  <xsl:template name="assessDignity">
    <xsl:param name="planet"/>
    <xsl:variable name="name" select="$planet/@name"/>
    <xsl:variable name="lon" select="$planet/Tropical/Lon"/>
    <xsl:variable name="signIdx" select="floor($lon div 30)"/>

    <western:Dignity>
      <western:Planet><xsl:value-of select="$name"/></western:Planet>
      <western:Sign>
        <xsl:call-template name="nthToken">
          <xsl:with-param name="list" select="'Aries,Taurus,Gemini,Cancer,Leo,Virgo,Libra,Scorpio,Sagittarius,Capricorn,Aquarius,Pisces'"/>
          <xsl:with-param name="n" select="$signIdx + 1"/>
        </xsl:call-template>
      </western:Sign>

      <western:Domicile>
        <xsl:call-template name="checkDomicile">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </western:Domicile>

      <western:Detriment>
        <xsl:call-template name="checkDetriment">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </western:Detriment>

      <western:Exaltation>
        <xsl:call-template name="checkExaltation">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </western:Exaltation>

      <western:Fall>
        <xsl:call-template name="checkFall">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </western:Fall>

      <western:State>
        <xsl:call-template name="determineState">
          <xsl:with-param name="planet" select="$name"/>
          <xsl:with-param name="signIdx" select="$signIdx"/>
        </xsl:call-template>
      </western:State>
    </western:Dignity>
  </xsl:template>

  <!-- ── Domicile check ────────────────────────────────────────────────── -->
  <xsl:template name="checkDomicile">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 4">true</xsl:when>
      <xsl:when test="$planet = 'Moon' and $signIdx = 3">true</xsl:when>
      <xsl:when test="$planet = 'Mercury' and ($signIdx = 2 or $signIdx = 5)">true</xsl:when>
      <xsl:when test="$planet = 'Venus' and ($signIdx = 1 or $signIdx = 6)">true</xsl:when>
      <xsl:when test="$planet = 'Mars' and ($signIdx = 0 or $signIdx = 7)">true</xsl:when>
      <xsl:when test="$planet = 'Jupiter' and ($signIdx = 8 or $signIdx = 11)">true</xsl:when>
      <xsl:when test="$planet = 'Saturn' and ($signIdx = 9 or $signIdx = 10)">true</xsl:when>
      <xsl:when test="$planet = 'Uranus' and $signIdx = 10">true</xsl:when>
      <xsl:when test="$planet = 'Neptune' and $signIdx = 11">true</xsl:when>
      <xsl:when test="$planet = 'Pluto' and $signIdx = 7">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Detriment check (opposite domicile) ───────────────────────────── -->
  <xsl:template name="checkDetriment">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 10">true</xsl:when>
      <xsl:when test="$planet = 'Moon' and $signIdx = 9">true</xsl:when>
      <xsl:when test="$planet = 'Mercury' and ($signIdx = 8 or $signIdx = 11)">true</xsl:when>
      <xsl:when test="$planet = 'Venus' and ($signIdx = 7 or $signIdx = 0)">true</xsl:when>
      <xsl:when test="$planet = 'Mars' and ($signIdx = 6 or $signIdx = 1)">true</xsl:when>
      <xsl:when test="$planet = 'Jupiter' and ($signIdx = 2 or $signIdx = 5)">true</xsl:when>
      <xsl:when test="$planet = 'Saturn' and ($signIdx = 3 or $signIdx = 4)">true</xsl:when>
      <xsl:when test="$planet = 'Uranus' and $signIdx = 4">true</xsl:when>
      <xsl:when test="$planet = 'Neptune' and $signIdx = 5">true</xsl:when>
      <xsl:when test="$planet = 'Pluto' and $signIdx = 1">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Exaltation check ──────────────────────────────────────────────── -->
  <xsl:template name="checkExaltation">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 0">true</xsl:when>
      <xsl:when test="$planet = 'Moon' and $signIdx = 1">true</xsl:when>
      <xsl:when test="$planet = 'Mercury' and $signIdx = 5">true</xsl:when>
      <xsl:when test="$planet = 'Venus' and $signIdx = 11">true</xsl:when>
      <xsl:when test="$planet = 'Mars' and $signIdx = 9">true</xsl:when>
      <xsl:when test="$planet = 'Jupiter' and $signIdx = 3">true</xsl:when>
      <xsl:when test="$planet = 'Saturn' and $signIdx = 6">true</xsl:when>
      <xsl:when test="$planet = 'Uranus' and $signIdx = 7">true</xsl:when>
      <xsl:when test="$planet = 'Neptune' and $signIdx = 3">true</xsl:when>
      <xsl:when test="$planet = 'Pluto' and $signIdx = 0">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── Fall check (opposite exaltation) ──────────────────────────────── -->
  <xsl:template name="checkFall">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:choose>
      <xsl:when test="$planet = 'Sun' and $signIdx = 6">true</xsl:when>
      <xsl:when test="$planet = 'Moon' and $signIdx = 7">true</xsl:when>
      <xsl:when test="$planet = 'Mercury' and $signIdx = 11">true</xsl:when>
      <xsl:when test="$planet = 'Venus' and $signIdx = 5">true</xsl:when>
      <xsl:when test="$planet = 'Mars' and $signIdx = 3">true</xsl:when>
      <xsl:when test="$planet = 'Jupiter' and $signIdx = 9">true</xsl:when>
      <xsl:when test="$planet = 'Saturn' and $signIdx = 0">true</xsl:when>
      <xsl:when test="$planet = 'Uranus' and $signIdx = 1">true</xsl:when>
      <xsl:when test="$planet = 'Neptune' and $signIdx = 9">true</xsl:when>
      <xsl:when test="$planet = 'Pluto' and $signIdx = 6">true</xsl:when>
      <xsl:otherwise>false</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

  <!-- ── State determination ──────────────────────────────────────────── -->
  <xsl:template name="determineState">
    <xsl:param name="planet"/>
    <xsl:param name="signIdx"/>
    <xsl:variable name="domicile">
      <xsl:call-template name="checkDomicile">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="exaltation">
      <xsl:call-template name="checkExaltation">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="detriment">
      <xsl:call-template name="checkDetriment">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:variable name="fall">
      <xsl:call-template name="checkFall">
        <xsl:with-param name="planet" select="$planet"/>
        <xsl:with-param name="signIdx" select="$signIdx"/>
      </xsl:call-template>
    </xsl:variable>
    <xsl:choose>
      <xsl:when test="$domicile = 'true'">domicile</xsl:when>
      <xsl:when test="$exaltation = 'true'">exaltation</xsl:when>
      <xsl:when test="$detriment = 'true'">detriment</xsl:when>
      <xsl:when test="$fall = 'true'">fall</xsl:when>
      <xsl:otherwise>peregrine</xsl:otherwise>
    </xsl:choose>
  </xsl:template>

</xsl:stylesheet>
